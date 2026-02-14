package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"google.golang.org/genai"
)

// GeminiClient wraps the Gemini SDK for text and multimodal generation.
type GeminiClient struct {
	client *genai.Client
	model  string
}

// NewGeminiClient creates a new Gemini API client.
func NewGeminiClient(ctx context.Context, apiKey, model string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &GeminiClient{client: client, model: model}, nil
}

// resolveModel returns the override model if non-empty, otherwise the default.
func (c *GeminiClient) resolveModel(override string) string {
	if override != "" {
		return override
	}
	return c.model
}

const maxRetries = 3

// isTransient returns true for errors that are worth retrying (network hiccups).
func isTransient(err error) bool {
	if err == nil {
		return false
	}
	// unexpected EOF, connection reset, timeout
	if err == io.ErrUnexpectedEOF {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	msg := err.Error()
	for _, substr := range []string{"unexpected EOF", "connection reset", "broken pipe", "UNAVAILABLE", "503", "429"} {
		if strings.Contains(msg, substr) {
			return true
		}
	}
	return false
}

// generateWithRetry wraps GenerateContent with retries for transient errors.
func (c *GeminiClient) generateWithRetry(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * 2 * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
		resp, err := c.client.Models.GenerateContent(ctx, model, contents, config)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isTransient(err) {
			return nil, err
		}
	}
	return nil, lastErr
}

// Generate calls the model with a system prompt and user message.
func (c *GeminiClient) Generate(ctx context.Context, systemPrompt, userMsg string, temp float32, modelOverrides ...string) (string, error) {
	var override string
	if len(modelOverrides) > 0 {
		override = modelOverrides[0]
	}

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(temp),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemPrompt)},
		},
	}

	resp, err := c.generateWithRetry(ctx, c.resolveModel(override), []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{genai.NewPartFromText(userMsg)},
		},
	}, config)
	if err != nil {
		return "", fmt.Errorf("gemini generate: %w", err)
	}

	text := resp.Text()
	return text, nil
}

// GenerateWithImage calls the model with a system prompt, user message, and image.
func (c *GeminiClient) GenerateWithImage(ctx context.Context, systemPrompt, userMsg string, img []byte, temp float32, modelOverrides ...string) (string, error) {
	var override string
	if len(modelOverrides) > 0 {
		override = modelOverrides[0]
	}

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(temp),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemPrompt)},
		},
	}

	resp, err := c.generateWithRetry(ctx, c.resolveModel(override), []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				genai.NewPartFromText(userMsg),
				genai.NewPartFromBytes(img, "image/png"),
			},
		},
	}, config)
	if err != nil {
		return "", fmt.Errorf("gemini generate with image: %w", err)
	}

	text := resp.Text()
	return text, nil
}

// GenerateWithThinking calls the model with thinking mode enabled.
// thinkingBudget controls the maximum thinking tokens (0 = model default, typically 8192).
func (c *GeminiClient) GenerateWithThinking(ctx context.Context, systemPrompt, userMsg string, temp float32, thinkingBudget int32, modelOverrides ...string) (string, error) {
	var override string
	if len(modelOverrides) > 0 {
		override = modelOverrides[0]
	}

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(temp),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemPrompt)},
		},
		ThinkingConfig: &genai.ThinkingConfig{
			ThinkingBudget: genai.Ptr(thinkingBudget),
		},
	}

	resp, err := c.generateWithRetry(ctx, c.resolveModel(override), []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{genai.NewPartFromText(userMsg)},
		},
	}, config)
	if err != nil {
		return "", fmt.Errorf("gemini generate with thinking: %w", err)
	}

	text := resp.Text()
	return text, nil
}

// ReviewImage calls the model with lower temperature for visual quality review.
func (c *GeminiClient) ReviewImage(ctx context.Context, systemPrompt, userMsg string, img []byte, modelOverrides ...string) (string, error) {
	return c.GenerateWithImage(ctx, systemPrompt, userMsg, img, 0.2, modelOverrides...)
}
