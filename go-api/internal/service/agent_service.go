package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/thesisviz/go-api/internal/agent"
	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/model"
	"github.com/thesisviz/go-api/internal/prompt"
	"github.com/thesisviz/go-api/internal/storage"
)

const (
	maxCompileRetries = 3
	maxReviewRounds   = 3
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
}

type GenerateRequest struct {
	ProjectID      string
	Format         string
	Prompt         string
	Language       string
	ColorScheme    string
	ThesisTitle    string
	ThesisAbstract string
	Model          string
}

type RefineRequest struct {
	GenerationID string
	Modification string
	Language     string
	ColorScheme  string
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

// Generate runs the full pipeline: generate → compile/fix → review/fix → explain.
func (s *AgentService) Generate(ctx context.Context, req GenerateRequest, pushFn func(ProgressMsg)) (*GenerateResult, error) {
	ag, ok := s.agents[req.Format]
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", req.Format)
	}

	opts := agent.AgentOpts{
		Language:       req.Language,
		ColorScheme:    req.ColorScheme,
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

	// === Phase 2: Compile + fix (for tikz/matplotlib, not mermaid) ===
	var imageBytes []byte
	var imageURL, imageKey string

	if req.Format != "mermaid" {
		pushFn(ProgressMsg{Type: "status", Phase: "compiling", Data: ProgressData{
			Message: "Compiling...", Progress: 30,
		}})

		for attempt := 1; attempt <= maxCompileRetries; attempt++ {
			renderResp, renderErr := s.renderSvc.RenderCode(ctx, RenderCodeRequest{
				Code:        code,
				Format:      req.Format,
				Language:    req.Language,
				ColorScheme: req.ColorScheme,
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
					s.markFailed(gen)
					return nil, fmt.Errorf("compilation failed: %s", errMsg)
				}

				pushFn(ProgressMsg{Type: "status", Phase: "fixing", Data: ProgressData{
					Message:  fmt.Sprintf("Compile error (attempt %d/%d), fixing...", attempt, maxCompileRetries),
					Progress: 30 + attempt*5,
					Round:    attempt,
				}})

				code, err = ag.Refine(ctx, code, "Compilation error: "+errMsg, opts)
				if err != nil {
					s.logger.Warn("refine failed", zap.Error(err))
					continue
				}
				continue
			}

			// Compilation succeeded
			imageURL = renderResp.ImageURL
			imageKey = renderResp.ImageKey
			// Download image bytes for review
			imageBytes, _ = s.storage.Download(ctx, imageKey)
			break
		}
	}

	// === Phase 3: Visual review + fix (only if we have an image) ===
	reviewPassed := false
	reviewRounds := 0

	if len(imageBytes) > 0 {
		for round := 1; round <= maxReviewRounds; round++ {
			reviewRounds = round
			pushFn(ProgressMsg{Type: "status", Phase: "reviewing", Data: ProgressData{
				Message:  fmt.Sprintf("Visual review round %d...", round),
				Progress: 50 + round*10,
				Round:    round,
				ImageURL: imageURL,
			}})

			reviewSys := prompt.ReviewSystem(req.Language)
			reviewUser := "Review the attached image for layout issues and content completeness."
			if req.Prompt != "" {
				reviewUser += "\n\nOriginal drawing prompt: " + req.Prompt
			}

			reviewRaw, reviewErr := s.llm.ReviewImage(ctx, reviewSys, reviewUser, imageBytes, req.Model)
			if reviewErr != nil {
				s.logger.Warn("review failed", zap.Error(reviewErr))
				reviewPassed = true // Skip review on error
				break
			}

			reviewJSON, parseErr := agent.ParseJSON(reviewRaw)
			if parseErr != nil {
				s.logger.Warn("review parse failed", zap.Error(parseErr))
				reviewPassed = true
				break
			}

			var reviewResult struct {
				Passed bool     `json:"passed"`
				Issues []string `json:"issues"`
			}
			if err := json.Unmarshal([]byte(reviewJSON), &reviewResult); err != nil {
				s.logger.Warn("review unmarshal failed", zap.Error(err))
				reviewPassed = true
				break
			}

			if reviewResult.Passed || len(reviewResult.Issues) == 0 {
				reviewPassed = true
				break
			}

			// Not passed — try to fix
			pushFn(ProgressMsg{Type: "preview", Phase: "fixing", Data: ProgressData{
				Message:  fmt.Sprintf("Issues found in round %d, fixing...", round),
				Round:    round,
				Issues:   reviewResult.Issues,
				ImageURL: imageURL,
			}})

			fixPrompt := prompt.ReviewFix(reviewResult.Issues, req.Language, req.Prompt)
			newCode, fixErr := ag.RefineWithImage(ctx, code, fixPrompt, imageBytes, opts)
			if fixErr != nil {
				s.logger.Warn("review fix failed", zap.Error(fixErr))
				break
			}
			code = newCode

			// Re-render after fix
			renderResp, renderErr := s.renderSvc.RenderCode(ctx, RenderCodeRequest{
				Code:        code,
				Format:      req.Format,
				Language:    req.Language,
				ColorScheme: req.ColorScheme,
			})
			if renderErr != nil || renderResp.Status == "error" {
				s.logger.Warn("re-render after fix failed")
				break
			}
			imageURL = renderResp.ImageURL
			imageKey = renderResp.ImageKey
			imageBytes, _ = s.storage.Download(ctx, imageKey)
		}
	} else if req.Format == "mermaid" {
		// Mermaid is rendered client-side, skip review
		reviewPassed = true
	}

	// === Phase 4: Code explanation ===
	pushFn(ProgressMsg{Type: "status", Phase: "explaining", Data: ProgressData{
		Message: "Generating explanation...", Progress: 85,
	}})

	explanationSys := prompt.Explanation(req.Format, req.Language)
	explanation, _ := s.llm.Generate(ctx, explanationSys, code, 0.4, req.Model)

	// === Phase 5: Save result ===
	codePtr := &code
	gen.Code = codePtr
	gen.Status = "success"
	if imageKey != "" {
		gen.ImageKey = &imageKey
	}
	if explanation != "" {
		gen.Explanation = &explanation
	}
	if reviewRounds > 0 {
		issuesJSON, _ := json.Marshal(map[string]interface{}{
			"review_passed": reviewPassed,
			"review_rounds": reviewRounds,
		})
		issuesStr := string(issuesJSON)
		gen.ReviewIssues = &issuesStr
	}
	_ = s.genSvc.Update(gen)

	result := &GenerateResult{
		GenerationID: gen.ID.String(),
		Code:         code,
		Format:       req.Format,
		Explanation:  explanation,
		ImageURL:     imageURL,
		ReviewPassed: reviewPassed,
		ReviewRounds: reviewRounds,
	}

	pushFn(ProgressMsg{Type: "result", Phase: "done", Data: ProgressData{
		Message:      "Generation complete",
		Progress:     100,
		GenerationID: gen.ID.String(),
		Code:         code,
		Format:       req.Format,
		Explanation:  explanation,
		ImageURL:     imageURL,
		ReviewPassed: reviewPassed,
		ReviewRounds: reviewRounds,
	}})

	return result, nil
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

	format := gen.Format
	ag, ok := s.agents[format]
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	_ = ag

	// Use the full generate pipeline with the refine as a new prompt
	result, err := s.Generate(ctx, GenerateRequest{
		ProjectID:   gen.ProjectID.String(),
		Format:      format,
		Prompt:      fmt.Sprintf("Modify this existing code:\n\n%s\n\nModification: %s", *gen.Code, req.Modification),
		Language:    req.Language,
		ColorScheme: req.ColorScheme,
		Model:       req.Model,
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

func (s *AgentService) markFailed(gen *model.Generation) {
	gen.Status = "failed"
	_ = s.genSvc.Update(gen)
}

// renderAndGetBytes is a helper to render and also download the image bytes.
func (s *AgentService) renderAndGetBytes(ctx context.Context, code, format, language, colorScheme string) (imageURL, imageKey string, imageBytes []byte, err error) {
	resp, err := s.renderSvc.RenderCode(ctx, RenderCodeRequest{
		Code:        code,
		Format:      format,
		Language:    language,
		ColorScheme: colorScheme,
	})
	if err != nil {
		return "", "", nil, err
	}
	if resp.Status == "error" {
		return "", "", nil, fmt.Errorf("render error: %s", resp.Error)
	}

	imageBytes, _ = s.storage.Download(ctx, resp.ImageKey)
	return resp.ImageURL, resp.ImageKey, imageBytes, nil
}

