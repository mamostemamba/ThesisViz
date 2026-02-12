package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/thesisviz/go-api/internal/agent"
	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/model"
	"github.com/thesisviz/go-api/internal/prompt"
	"github.com/thesisviz/go-api/internal/storage"
	"github.com/thesisviz/go-api/pkg/colorscheme"
)

const (
	maxCompileRetries  = 3
	maxRerolls         = 3
	maxFixRounds       = 5
	minScoreToFixPhase = 4.0 // bestScore >= 此值时跳出重画，进入修复润色
)

// ProgressMsg is the WebSocket message sent to clients during generation.
type ProgressMsg struct {
	Type  string       `json:"type"`  // status, preview, result, error
	Phase string       `json:"phase"` // generating, compiling, reviewing, fixing, explaining, done
	Data  ProgressData `json:"data"`
}

type ProgressData struct {
	Message      string   `json:"message,omitempty"`
	Progress     int      `json:"progress,omitempty"`
	Round        int      `json:"round,omitempty"`
	ImageURL     string   `json:"image_url,omitempty"`
	Issues       []string `json:"issues,omitempty"`
	GenerationID string   `json:"generation_id,omitempty"`
	Code         string   `json:"code,omitempty"`
	Format       string   `json:"format,omitempty"`
	Explanation  string   `json:"explanation,omitempty"`
	ReviewPassed bool     `json:"review_passed,omitempty"`
	ReviewRounds int      `json:"review_rounds,omitempty"`
	Critique     string   `json:"critique,omitempty"`
	Score        float64  `json:"score,omitempty"`
}

type GenerateRequest struct {
	ProjectID      string
	Format         string
	Prompt         string
	Language       string
	ColorScheme    string
	CustomColors   *colorscheme.CustomColors
	ThesisTitle    string
	ThesisAbstract string
	Model          string
}

type RefineRequest struct {
	GenerationID string
	Modification string
	Language     string
	ColorScheme  string
	CustomColors *colorscheme.CustomColors
	Model        string
}

type AnalyzeRequest struct {
	Text           string
	Language       string
	ThesisTitle    string
	ThesisAbstract string
	Model          string
}

type GenerateResult struct {
	GenerationID string
	Code         string
	Format       string
	Explanation  string
	ImageURL     string
	ReviewPassed bool
	ReviewRounds int
}

type AgentService struct {
	llm       *llm.GeminiClient
	renderSvc *RenderService
	genSvc    *GenerationService
	storage   *storage.MinIOStorage
	agents    map[string]agent.Agent
	router    *agent.RouterAgent
	logger    *zap.Logger
}

func NewAgentService(
	llmClient *llm.GeminiClient,
	renderSvc *RenderService,
	genSvc *GenerationService,
	store *storage.MinIOStorage,
	agents map[string]agent.Agent,
	router *agent.RouterAgent,
	logger *zap.Logger,
) *AgentService {
	return &AgentService{
		llm:       llmClient,
		renderSvc: renderSvc,
		genSvc:    genSvc,
		storage:   store,
		agents:    agents,
		router:    router,
		logger:    logger,
	}
}

// Analyze runs the router agent to recommend figures from thesis text.
func (s *AgentService) Analyze(ctx context.Context, req AnalyzeRequest) ([]agent.Recommendation, error) {
	opts := agent.AgentOpts{
		Language:       req.Language,
		ThesisTitle:    req.ThesisTitle,
		ThesisAbstract: req.ThesisAbstract,
		Model:          req.Model,
	}
	return s.router.Analyze(ctx, req.Text, opts)
}

// reviewPhaseResult holds the outcome of the visual review phase.
type reviewPhaseResult struct {
	Code         string
	ImageURL     string
	ImageKey     string
	ReviewPassed bool
	ReviewRounds int
	Score        float64
	Issues       []string
	Critique     string
}

// Generate runs the full pipeline: generate -> compile/fix -> review/fix -> explain.
func (s *AgentService) Generate(ctx context.Context, req GenerateRequest, pushFn func(ProgressMsg)) (*GenerateResult, error) {
	ag, ok := s.agents[req.Format]
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", req.Format)
	}

	opts := agent.AgentOpts{
		Language:       req.Language,
		ColorScheme:    req.ColorScheme,
		CustomColors:   req.CustomColors,
		ThesisTitle:    req.ThesisTitle,
		ThesisAbstract: req.ThesisAbstract,
		Model:          req.Model,
	}

	// Create generation record
	gen := &model.Generation{
		Format: req.Format,
		Prompt: req.Prompt,
		Status: "processing",
	}
	if req.ProjectID != "" {
		pid, err := uuid.Parse(req.ProjectID)
		if err == nil {
			gen.ProjectID = pid
		}
	}
	if err := s.genSvc.Create(gen); err != nil {
		s.logger.Warn("failed to create generation record", zap.Error(err))
	}

	// === Phase 1: Generate code ===
	pushFn(ProgressMsg{Type: "status", Phase: "generating", Data: ProgressData{
		Message: "Generating code...", Progress: 10,
	}})

	code, err := ag.Generate(ctx, req.Prompt, opts)
	if err != nil {
		pushFn(ProgressMsg{Type: "error", Phase: "generating", Data: ProgressData{Message: err.Error()}})
		s.markFailed(gen)
		return nil, err
	}

	pushFn(ProgressMsg{Type: "preview", Phase: "generating", Data: ProgressData{
		Message: "代码生成完成", Progress: 20, Code: code,
	}})

	// === Phase 2: Compile + fix ===
	code, imageURL, imageKey, imageBytes, err := s.compileWithRetries(ctx, ag, code, req, opts, pushFn)
	if err != nil {
		s.markFailed(gen)
		return nil, err
	}

	// === Phase 3: Visual review ===
	rr := s.visualReview(ctx, ag, code, imageURL, imageKey, imageBytes, req, opts, pushFn)
	code = rr.Code
	imageURL = rr.ImageURL
	imageKey = rr.ImageKey

	// === Phase 4: Save result ===
	codePtr := &code
	gen.Code = codePtr
	gen.Status = "success"
	if imageKey != "" {
		gen.ImageKey = &imageKey
	}
	if rr.ReviewRounds > 0 {
		issuesJSON, _ := json.Marshal(map[string]interface{}{
			"review_passed": rr.ReviewPassed,
			"review_rounds": rr.ReviewRounds,
		})
		issuesStr := string(issuesJSON)
		gen.ReviewIssues = &issuesStr
	}
	_ = s.genSvc.Update(gen)

	result := &GenerateResult{
		GenerationID: gen.ID.String(),
		Code:         code,
		Format:       req.Format,
		ImageURL:     imageURL,
		ReviewPassed: rr.ReviewPassed,
		ReviewRounds: rr.ReviewRounds,
	}

	pushFn(ProgressMsg{Type: "result", Phase: "done", Data: ProgressData{
		Message:      "Generation complete",
		Progress:     100,
		GenerationID: gen.ID.String(),
		Code:         code,
		Format:       req.Format,
		ImageURL:     imageURL,
		ReviewPassed: rr.ReviewPassed,
		ReviewRounds: rr.ReviewRounds,
		Critique:     rr.Critique,
		Issues:       rr.Issues,
		Score:        rr.Score,
	}})

	return result, nil
}

// compileWithRetries handles Phase 2: compile the code and retry with LLM fixes on failure.
// For mermaid format (client-side rendering), it returns immediately with no image.
func (s *AgentService) compileWithRetries(
	ctx context.Context, ag agent.Agent, code string,
	req GenerateRequest, opts agent.AgentOpts,
	pushFn func(ProgressMsg),
) (finalCode string, imageURL, imageKey string, imageBytes []byte, err error) {
	if req.Format == "mermaid" {
		return code, "", "", nil, nil
	}

	pushFn(ProgressMsg{Type: "status", Phase: "compiling", Data: ProgressData{
		Message: "Compiling...", Progress: 30,
	}})

	for attempt := 1; attempt <= maxCompileRetries; attempt++ {
		select {
		case <-ctx.Done():
			return code, "", "", nil, ctx.Err()
		default:
		}

		renderResp, renderErr := s.renderSvc.RenderCode(ctx, RenderCodeRequest{
			Code:         code,
			Format:       req.Format,
			Language:     req.Language,
			ColorScheme:  req.ColorScheme,
			CustomColors: req.CustomColors,
		})
		if renderErr != nil || renderResp.Status == "error" {
			errMsg := "unknown render error"
			if renderErr != nil {
				errMsg = renderErr.Error()
			} else if renderResp.Error != "" {
				errMsg = renderResp.Error
			}

			if attempt >= maxCompileRetries {
				pushFn(ProgressMsg{Type: "error", Phase: "compiling", Data: ProgressData{
					Message: fmt.Sprintf("Compilation failed after %d attempts: %s", maxCompileRetries, errMsg),
				}})
				return code, "", "", nil, fmt.Errorf("compilation failed: %s", errMsg)
			}

			pushFn(ProgressMsg{Type: "status", Phase: "compiling", Data: ProgressData{
				Message:  fmt.Sprintf("编译出错（第 %d/%d 次），AI 正在修复代码...", attempt, maxCompileRetries),
				Progress: 30 + attempt*5,
				Round:    attempt,
			}})

			refined, refErr := ag.Refine(ctx, code, "Compilation error: "+errMsg, opts)
			if refErr != nil {
				s.logger.Warn("refine failed", zap.Error(refErr))
				continue
			}
			code = refined

			pushFn(ProgressMsg{Type: "status", Phase: "compiling", Data: ProgressData{
				Message:  fmt.Sprintf("代码已修复，正在重新编译（第 %d 次）...", attempt+1),
				Progress: 30 + attempt*5 + 3,
				Round:    attempt,
			}})
			continue
		}

		// Compilation succeeded
		imageURL = renderResp.ImageURL
		imageKey = renderResp.ImageKey
		if dlBytes, dlErr := s.storage.Download(ctx, imageKey); dlErr != nil {
			s.logger.Warn("image download failed after compile", zap.Error(dlErr))
		} else {
			imageBytes = dlBytes
		}

		pushFn(ProgressMsg{Type: "preview", Phase: "compiling", Data: ProgressData{
			Message: "编译成功", Progress: 45, Code: code, ImageURL: imageURL,
		}})
		return code, imageURL, imageKey, imageBytes, nil
	}

	return code, "", "", nil, fmt.Errorf("compilation failed: exhausted retries")
}

// reviewIsPassed checks if a review result means "passed".
func reviewIsPassed(rev *reviewOutput) bool {
	if rev.Passed || rev.Score >= 9 {
		return true
	}
	if !rev.Passed && len(rev.Issues) == 0 {
		return true
	}
	return false
}

// visualReview handles Phase 3: initial review -> reroll loop -> fix loop.
func (s *AgentService) visualReview(
	ctx context.Context, ag agent.Agent,
	code, imageURL, imageKey string, imageBytes []byte,
	req GenerateRequest, opts agent.AgentOpts,
	pushFn func(ProgressMsg),
) reviewPhaseResult {
	result := reviewPhaseResult{
		Code:     code,
		ImageURL: imageURL,
		ImageKey: imageKey,
	}

	if req.Format == "mermaid" {
		result.ReviewPassed = true
		return result
	}

	if len(imageBytes) == 0 {
		return result
	}

	// Best version tracking (across initial + rerolls)
	bestCode := code
	bestImageURL := imageURL
	bestImageKey := imageKey
	bestImageBytes := imageBytes
	bestScore := float64(0)
	bestIssues := []string{}
	bestCritique := ""

	updateBest := func(c, url, key string, img []byte, rev *reviewOutput) {
		if rev.Score > bestScore {
			bestCode = c
			bestImageURL = url
			bestImageKey = key
			bestImageBytes = img
			bestScore = rev.Score
			bestIssues = rev.Issues
			bestCritique = rev.Critique
		}
	}

	// --- 3a. Initial review ---
	pushFn(ProgressMsg{Type: "status", Phase: "reviewing", Data: ProgressData{
		Message: "视觉审查中...", Progress: 50, ImageURL: imageURL,
	}})
	result.ReviewRounds++

	rev, revErr := s.doReview(ctx, imageBytes, req.Prompt, req.Language, req.Model)
	if revErr != nil {
		s.logger.Warn("initial review failed", zap.Error(revErr))
		result.Critique = fmt.Sprintf("审查调用失败: %s", revErr.Error())
	} else {
		s.logger.Info("initial review", zap.Float64("score", rev.Score), zap.Bool("passed", rev.Passed))
		result.Score = rev.Score
		result.Issues = rev.Issues
		result.Critique = rev.Critique
		updateBest(code, imageURL, imageKey, imageBytes, rev)

		if reviewIsPassed(rev) {
			result.ReviewPassed = true
			pushFn(ProgressMsg{Type: "preview", Phase: "reviewing", Data: ProgressData{
				Message: fmt.Sprintf("审查通过 (%.0f/10)", rev.Score), ImageURL: imageURL,
				Critique: rev.Critique, Issues: rev.Issues, Score: rev.Score,
			}})
			result.Code = code
			return result
		}

		pushFn(ProgressMsg{Type: "preview", Phase: "reviewing", Data: ProgressData{
			Message: fmt.Sprintf("初始审查 %.0f/10，进入重画阶段", rev.Score), ImageURL: imageURL,
			Critique: rev.Critique, Issues: rev.Issues, Score: rev.Score,
		}})
	}

	// --- 3b. Reroll loop (skip entirely if score already sufficient) ---
	if bestScore >= minScoreToFixPhase {
		s.logger.Info("score sufficient, skipping reroll phase",
			zap.Float64("best_score", bestScore),
		)
	} else {
		for reroll := 1; reroll <= maxRerolls; reroll++ {
			select {
			case <-ctx.Done():
				s.logger.Warn("context cancelled during reroll", zap.Error(ctx.Err()))
				result.Code = bestCode
				result.ImageURL = bestImageURL
				result.ImageKey = bestImageKey
				return result
			default:
			}

			// Check again after each reroll in case bestScore improved enough
			if bestScore >= minScoreToFixPhase {
				s.logger.Info("best score now sufficient, stopping rerolls",
					zap.Float64("best_score", bestScore),
					zap.Int("rerolls_done", reroll-1),
				)
				break
			}

		negativeHint := fmt.Sprintf(
			"\n\nIMPORTANT: A previous attempt scored %.0f/10 due to: %s. "+
				"You MUST avoid these problems. Use a completely different layout strategy.",
			result.Score, strings.Join(result.Issues, "; "),
		)

		pushFn(ProgressMsg{Type: "status", Phase: "rerolling", Data: ProgressData{
			Message:  fmt.Sprintf("重新生成 (%d/%d)...", reroll, maxRerolls),
			Progress: 50 + reroll*5,
			Round:    reroll,
			Score:    result.Score,
		}})

		s.logger.Info("triggering reroll",
			zap.Int("reroll", reroll),
			zap.Float64("prev_score", result.Score),
		)

		newCode, genErr := ag.Generate(ctx, req.Prompt+negativeHint, opts)
		if genErr != nil {
			s.logger.Warn("reroll generation failed", zap.Int("reroll", reroll), zap.Error(genErr))
			pushFn(ProgressMsg{Type: "preview", Phase: "rerolling", Data: ProgressData{
				Message: fmt.Sprintf("第 %d 次重画生成失败: %s", reroll, genErr.Error()),
				Round:   reroll,
			}})
			continue
		}

		pushFn(ProgressMsg{Type: "preview", Phase: "rerolling", Data: ProgressData{
			Message: fmt.Sprintf("第 %d 次重画生成完成，编译中...", reroll), Progress: 52 + reroll*5,
			Round: reroll, Code: newCode,
		}})

		rURL, rKey, rBytes, renderErr := s.renderAndGetBytes(ctx, newCode, req.Format, req.Language, req.ColorScheme, req.CustomColors)
		if renderErr != nil {
			s.logger.Warn("reroll render failed", zap.Int("reroll", reroll), zap.Error(renderErr))
			pushFn(ProgressMsg{Type: "preview", Phase: "rerolling", Data: ProgressData{
				Message: fmt.Sprintf("第 %d 次重画编译失败: %s", reroll, renderErr.Error()),
				Round:   reroll,
			}})
			continue
		}

		// Review the rerolled version
		result.ReviewRounds++
		rev, revErr := s.doReview(ctx, rBytes, req.Prompt, req.Language, req.Model)
		if revErr != nil {
			s.logger.Warn("reroll review failed", zap.Int("reroll", reroll), zap.Error(revErr))
			pushFn(ProgressMsg{Type: "preview", Phase: "rerolling", Data: ProgressData{
				Message:  fmt.Sprintf("第 %d 次重画审查失败: %s", reroll, revErr.Error()),
				Round:    reroll,
				ImageURL: rURL,
			}})
			continue
		}

		result.Score = rev.Score
		result.Issues = rev.Issues
		result.Critique = rev.Critique
		updateBest(newCode, rURL, rKey, rBytes, rev)

		s.logger.Info("reroll review result",
			zap.Int("reroll", reroll),
			zap.Float64("score", rev.Score),
			zap.Float64("best_score", bestScore),
			zap.Bool("passed", rev.Passed),
		)

		if reviewIsPassed(rev) {
			result.ReviewPassed = true
			result.Code = newCode
			result.ImageURL = rURL
			result.ImageKey = rKey
			pushFn(ProgressMsg{Type: "preview", Phase: "rerolling", Data: ProgressData{
				Message:  fmt.Sprintf("第 %d 次重画通过 (%.0f/10)", reroll, rev.Score),
				Round:    reroll,
				ImageURL: rURL, Code: newCode,
				Critique: rev.Critique, Issues: rev.Issues, Score: rev.Score,
			}})
			return result
		}

		pushFn(ProgressMsg{Type: "preview", Phase: "rerolling", Data: ProgressData{
			Message:  fmt.Sprintf("第 %d 次重画 %.0f/10，继续...", reroll, rev.Score),
			Round:    reroll,
			ImageURL: rURL, Code: newCode,
			Critique: rev.Critique, Issues: rev.Issues, Score: rev.Score,
		}})
		}
	} // end else (reroll loop)

	// --- 3c. Switch to best version for fix phase ---
	code = bestCode
	imageURL = bestImageURL
	imageKey = bestImageKey
	imageBytes = bestImageBytes
	result.Score = bestScore
	result.Issues = bestIssues
	result.Critique = bestCritique

	s.logger.Info("using best version for fix phase",
		zap.Float64("best_score", bestScore),
	)

	pushFn(ProgressMsg{Type: "status", Phase: "fixing", Data: ProgressData{
		Message:  fmt.Sprintf("选择最佳版本 (%.0f/10) 进行修复润色...", bestScore),
		Progress: 70,
		Score:    bestScore, ImageURL: imageURL, Code: code,
		Critique: bestCritique, Issues: bestIssues,
	}})

	// --- Fix loop ---
	for fix := 1; fix <= maxFixRounds; fix++ {
		select {
		case <-ctx.Done():
			s.logger.Warn("context cancelled during fix", zap.Error(ctx.Err()))
			result.Code = code
			result.ImageURL = imageURL
			result.ImageKey = imageKey
			return result
		default:
		}

		pushFn(ProgressMsg{Type: "preview", Phase: "fixing", Data: ProgressData{
			Message:  fmt.Sprintf("第 %d/%d 轮修复 (%.0f/10)...", fix, maxFixRounds, result.Score),
			Progress: 70 + fix*3,
			Round:    fix,
			Issues:   result.Issues, ImageURL: imageURL, Code: code,
			Critique: result.Critique, Score: result.Score,
		}})

		fixPrompt := prompt.ReviewFix(result.Issues, result.Score, req.Language, req.Prompt)
		newCode, fixErr := ag.RefineWithImage(ctx, code, fixPrompt, imageBytes, opts)
		if fixErr != nil {
			s.logger.Warn("fix failed", zap.Int("fix", fix), zap.Error(fixErr))
			pushFn(ProgressMsg{Type: "preview", Phase: "fixing", Data: ProgressData{
				Message:  fmt.Sprintf("第 %d 轮修复失败: %s", fix, fixErr.Error()),
				Round:    fix,
				Critique: fmt.Sprintf("修复代码生成失败: %s", fixErr.Error()),
			}})
			continue
		}
		code = newCode

		// Re-render
		fURL, fKey, fBytes, renderErr := s.renderAndGetBytes(ctx, code, req.Format, req.Language, req.ColorScheme, req.CustomColors)
		if renderErr != nil {
			s.logger.Warn("fix render failed", zap.Int("fix", fix), zap.Error(renderErr))
			pushFn(ProgressMsg{Type: "preview", Phase: "fixing", Data: ProgressData{
				Message:  fmt.Sprintf("第 %d 轮修复后编译失败: %s", fix, renderErr.Error()),
				Round:    fix,
				Critique: fmt.Sprintf("修复后渲染失败: %s", renderErr.Error()),
			}})
			continue
		}
		imageURL = fURL
		imageKey = fKey
		imageBytes = fBytes

		// Re-review
		result.ReviewRounds++
		rev, revErr := s.doReview(ctx, imageBytes, req.Prompt, req.Language, req.Model)
		if revErr != nil {
			s.logger.Warn("fix review failed", zap.Int("fix", fix), zap.Error(revErr))
			pushFn(ProgressMsg{Type: "preview", Phase: "fixing", Data: ProgressData{
				Message:  fmt.Sprintf("第 %d 轮修复后审查失败: %s", fix, revErr.Error()),
				Round:    fix,
				ImageURL: imageURL, Code: code,
			}})
			continue
		}

		result.Score = rev.Score
		result.Issues = rev.Issues
		result.Critique = rev.Critique

		s.logger.Info("fix review result",
			zap.Int("fix", fix),
			zap.Float64("score", rev.Score),
			zap.Bool("passed", rev.Passed),
		)

		if reviewIsPassed(rev) {
			result.ReviewPassed = true
			pushFn(ProgressMsg{Type: "preview", Phase: "reviewing", Data: ProgressData{
				Message:  fmt.Sprintf("修复后审查通过 (%.0f/10)", rev.Score),
				Round:    fix,
				ImageURL: imageURL, Code: code,
				Critique: rev.Critique, Issues: rev.Issues, Score: rev.Score,
			}})
			break
		}

		pushFn(ProgressMsg{Type: "preview", Phase: "reviewing", Data: ProgressData{
			Message:  fmt.Sprintf("第 %d 轮修复后 %.0f/10，继续...", fix, rev.Score),
			Round:    fix,
			ImageURL: imageURL, Code: code,
			Critique: rev.Critique, Issues: rev.Issues, Score: rev.Score,
		}})
	}

	result.Code = code
	result.ImageURL = imageURL
	result.ImageKey = imageKey
	return result
}

// Refine modifies an existing generation.
func (s *AgentService) Refine(ctx context.Context, req RefineRequest, pushFn func(ProgressMsg)) (*GenerateResult, error) {
	genID, err := uuid.Parse(req.GenerationID)
	if err != nil {
		return nil, fmt.Errorf("invalid generation id: %w", err)
	}

	gen, err := s.genSvc.GetByID(genID)
	if err != nil {
		return nil, fmt.Errorf("generation not found: %w", err)
	}

	if gen.Code == nil {
		return nil, fmt.Errorf("generation has no code to refine")
	}

	// Use the full generate pipeline with the refine as a new prompt
	result, err := s.Generate(ctx, GenerateRequest{
		ProjectID:    gen.ProjectID.String(),
		Format:       gen.Format,
		Prompt:       fmt.Sprintf("Modify this existing code:\n\n%s\n\nModification: %s", *gen.Code, req.Modification),
		Language:     req.Language,
		ColorScheme:  req.ColorScheme,
		CustomColors: req.CustomColors,
		Model:        req.Model,
	}, pushFn)
	if err != nil {
		return nil, err
	}

	// Set parent ID for chain tracking
	if result != nil && result.GenerationID != "" {
		childID, parseErr := uuid.Parse(result.GenerationID)
		if parseErr == nil {
			childGen, getErr := s.genSvc.GetByID(childID)
			if getErr == nil {
				childGen.ParentID = &gen.ID
				_ = s.genSvc.Update(childGen)
			}
		}
	}

	return result, nil
}

// reviewOutput holds parsed visual review results.
type reviewOutput struct {
	Passed   bool
	Score    float64
	Issues   []string
	Critique string
}

// doReview performs a single visual review call and returns parsed results.
func (s *AgentService) doReview(ctx context.Context, imageBytes []byte, drawingPrompt, language, mdl string) (*reviewOutput, error) {
	reviewSys := prompt.ReviewSystem(language)
	reviewUser := "Review the attached image for layout issues and content completeness."
	if drawingPrompt != "" {
		reviewUser += "\n\nOriginal drawing prompt: " + drawingPrompt
	}

	reviewRaw, err := s.llm.ReviewImage(ctx, reviewSys, reviewUser, imageBytes, mdl)
	if err != nil {
		return nil, fmt.Errorf("review call: %w", err)
	}

	s.logger.Info("review raw response", zap.String("raw", reviewRaw))

	reviewJSON, err := agent.ParseJSON(reviewRaw)
	if err != nil {
		return nil, fmt.Errorf("review parse: %w", err)
	}

	var result struct {
		Passed   bool     `json:"passed"`
		Score    float64  `json:"score"`
		Issues   []string `json:"issues"`
		Critique string   `json:"critique"`
	}
	jsonStr := strings.TrimSpace(reviewJSON)
	if len(jsonStr) > 0 && jsonStr[0] == '[' {
		var arr []json.RawMessage
		if err := json.Unmarshal([]byte(jsonStr), &arr); err == nil && len(arr) > 0 {
			jsonStr = string(arr[0])
		}
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("review unmarshal: %w", err)
	}

	return &reviewOutput{
		Passed:   result.Passed,
		Score:    result.Score,
		Issues:   result.Issues,
		Critique: result.Critique,
	}, nil
}

func (s *AgentService) markFailed(gen *model.Generation) {
	gen.Status = "failed"
	_ = s.genSvc.Update(gen)
}

// renderAndGetBytes is a helper to render and also download the image bytes.
func (s *AgentService) renderAndGetBytes(ctx context.Context, code, format, language, colorScheme string, customColors *colorscheme.CustomColors) (imageURL, imageKey string, imageBytes []byte, err error) {
	resp, err := s.renderSvc.RenderCode(ctx, RenderCodeRequest{
		Code:         code,
		Format:       format,
		Language:     language,
		ColorScheme:  colorScheme,
		CustomColors: customColors,
	})
	if err != nil {
		return "", "", nil, err
	}
	if resp.Status == "error" {
		return "", "", nil, fmt.Errorf("render error: %s", resp.Error)
	}

	if dlBytes, dlErr := s.storage.Download(ctx, resp.ImageKey); dlErr != nil {
		s.logger.Warn("image download failed", zap.Error(dlErr))
	} else {
		imageBytes = dlBytes
	}
	return resp.ImageURL, resp.ImageKey, imageBytes, nil
}
