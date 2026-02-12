package renderer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MatplotlibRenderer struct {
	sidecarURL string
	client     *http.Client
}

func NewMatplotlibRenderer(sidecarURL string) *MatplotlibRenderer {
	return &MatplotlibRenderer{
		sidecarURL: sidecarURL,
		client:     &http.Client{Timeout: 120 * time.Second},
	}
}

type matplotlibRequest struct {
	Code    string `json:"code"`
	Timeout int    `json:"timeout"`
}

type matplotlibResponse struct {
	Status string `json:"status"`
	Image  string `json:"image,omitempty"` // base64 PNG
	Error  string `json:"error,omitempty"`
}

func (r *MatplotlibRenderer) Render(ctx context.Context, code string, opts RenderOpts) (*RenderResult, error) {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 60
	}

	reqBody := matplotlibRequest{
		Code:    code,
		Timeout: timeout,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := r.sidecarURL + "/render/matplotlib"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sidecar request failed: %w", err)
	}
	defer resp.Body.Close()

	var result matplotlibResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode sidecar response: %w", err)
	}

	if result.Status != "ok" {
		return nil, fmt.Errorf("matplotlib render error: %s", result.Error)
	}

	imgBytes, err := base64.StdEncoding.DecodeString(result.Image)
	if err != nil {
		return nil, fmt.Errorf("decode base64 image: %w", err)
	}

	return &RenderResult{
		ImageBytes: imgBytes,
		Format:     "png",
	}, nil
}
