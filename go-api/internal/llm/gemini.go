package llm

import (
	"context"
	"fmt"

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

// Generate calls the model with a system prompt and user message.
func (c *GeminiClient) Generate(ctx context.Context, systemPrompt, userMsg string, temp float32) (string, error) {
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(temp),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemPrompt)},
		},
	}

	resp, err := c.client.Models.GenerateContent(ctx, c.model, []*genai.Content{
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
func (c *GeminiClient) GenerateWithImage(ctx context.Context, systemPrompt, userMsg string, img []byte, temp float32) (string, error) {
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(temp),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemPrompt)},
		},
	}

	resp, err := c.client.Models.GenerateContent(ctx, c.model, []*genai.Content{
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

// ReviewImage calls the model with lower temperature for visual quality review.
func (c *GeminiClient) ReviewImage(ctx context.Context, systemPrompt, userMsg string, img []byte) (string, error) {
	return c.GenerateWithImage(ctx, systemPrompt, userMsg, img, 0.2)
}
