package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AI is a minimal OpenAI-compatible chat completions client. Configure with
// DRIFT_API_KEY (required), DRIFT_BASE_URL and DRIFT_MODEL (optional).
type AI struct {
	Key    string
	Base   string
	Model  string
	Client *http.Client
}

// NewAI returns nil when AI whisper mode is not configured.
func NewAI() *AI {
	key := os.Getenv("DRIFT_API_KEY")
	if key == "" {
		return nil
	}
	base := os.Getenv("DRIFT_BASE_URL")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	base = strings.TrimRight(base, "/")
	model := os.Getenv("DRIFT_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &AI{
		Key:    key,
		Base:   base,
		Model:  model,
		Client: &http.Client{Timeout: 6 * time.Second},
	}
}

// TimeOfDayName maps a local hour to the word used in AI prompts.
func TimeOfDayName(hour float64) string {
	switch {
	case hour >= 5 && hour < 8:
		return "dawn"
	case hour >= 8 && hour < 17:
		return "day"
	case hour >= 17 && hour < 20.5:
		return "dusk"
	default:
		return "night"
	}
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
	// GLM-family models accept a `thinking` switch; harmless elsewhere only
	// when the model name says so, because some servers reject unknown fields.
	Thinking *thinkingSwitch `json:"thinking,omitempty"`
}

type thinkingSwitch struct {
	Type string `json:"type"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"message"`
	} `json:"choices"`
}

// Fetch requests one short whisper line. The prompt asks for max 12 words;
// anything unusable comes back as an error so callers fall back cleanly.
func (a *AI) Fetch(timeOfDay string) (string, error) {
	prompt := fmt.Sprintf(
		"Write one short, gentle, calming line (max 12 words) fitting a quiet moonlit pond at %s. No emojis. No quotes.",
		timeOfDay,
	)
	body := chatRequest{
		Model: a.Model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   40,
		Temperature: 0.9,
	}
	if strings.Contains(strings.ToLower(a.Model), "glm") {
		body.Thinking = &thinkingSwitch{Type: "disabled"}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.Base+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Key)

	resp, err := a.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("drift: whisper API status %d", resp.StatusCode)
	}
	var cr chatResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&cr); err != nil {
		return "", err
	}
	if len(cr.Choices) == 0 {
		return "", errors.New("drift: whisper API returned no choices")
	}
	return Sanitize(cr.Choices[0].Message.Content), nil
}
