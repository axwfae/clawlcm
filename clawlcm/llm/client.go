package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/axwfae/clawlcm/logger"
)

const (
	defaultTimeoutMs   = 60000
	defaultTemperature = 0.3
	maxTokensLeaf      = 2000
	maxTokensCondensed = 1500
)

type Client struct {
	Model      string
	Provider   string
	APIKey     string
	BaseURL    string
	TimeoutMs  int
	httpClient *http.Client
	logger     logger.Logger
}

func NewClient(model, provider, apiKey, baseURL string, timeoutMs int, log logger.Logger) *Client {
	if timeoutMs == 0 {
		timeoutMs = defaultTimeoutMs
	}
	return &Client{
		Model:      model,
		Provider:   provider,
		APIKey:     apiKey,
		BaseURL:    baseURL,
		TimeoutMs:  timeoutMs,
		httpClient: &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
		logger:     log,
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Response struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Usage struct {
	InputTokens  int `json:"prompt_tokens"`
	OutputTokens int `json:"completion_tokens"`
}

func (c *Client) Complete(ctx context.Context, messages []Message, maxTokens int) (string, error) {
	if c.BaseURL == "" {
		c.BaseURL = c.getDefaultBaseURL()
	}

	reqBody := Request{
		Model:       c.Model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: defaultTemperature,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	c.logger.Debug(fmt.Sprintf("LLM request to %s with model %s", c.BaseURL, c.Model))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices")
	}

	return response.Choices[0].Message.Content, nil
}

func (c *Client) getDefaultBaseURL() string {
	return "https://api.openai.com"
}

type Summarizer struct {
	client *Client
	log    logger.Logger
}

func NewSummarizer(client *Client, log logger.Logger) *Summarizer {
	return &Summarizer{client: client, log: log}
}

func stripAuthErrors(text string) string {
	lines := strings.Split(text, "\n")
	var filtered []string
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "http") && strings.Contains(lower, "error") {
			continue
		}
		if (strings.Contains(lower, "401") || strings.Contains(lower, "403") || strings.Contains(lower, "api key")) &&
			!strings.Contains(lower, "discussing") && !strings.Contains(lower, "about") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

func (s *Summarizer) SummarizeLeaf(messages []Message) (string, error) {
	systemPrompt := `You are a context-compaction summarization engine. 
Convert the following conversation into a concise summary that preserves key information, decisions, and context.
Focus on: main topics, important details, conclusions, and any actionable items.

Output only the summary text, no explanations or meta-commentary.`

	msgs := append([]Message{{Role: "system", Content: systemPrompt}}, messages...)

	result, err := s.client.Complete(context.Background(), msgs, maxTokensLeaf)
	if err != nil {
		s.log.Error(fmt.Sprintf("Leaf summarization failed: %v", err))
		return "", err
	}

	s.log.Info(fmt.Sprintf("Leaf summary created, length: %d chars", len(result)))
	return result, nil
}

func (s *Summarizer) SummarizeCondensed(summaries []Message) (string, error) {
	systemPrompt := `You are a high-level context-compaction engine.
Synthesize multiple conversation summaries into a coherent, unified summary.
Preserve the most important information across all summaries.

Output only the synthesized summary text, no explanations.`

	msgs := append([]Message{{Role: "system", Content: systemPrompt}}, summaries...)

	result, err := s.client.Complete(context.Background(), msgs, maxTokensCondensed)
	if err != nil {
		s.log.Error(fmt.Sprintf("Condensed summarization failed: %v", err))
		return "", err
	}

	s.log.Info(fmt.Sprintf("Condensed summary created, length: %d chars", len(result)))
	return result, nil
}

func (s *Summarizer) ExpandQuery(question, ctxContent string, maxTokens int) (string, int, error) {
	systemPrompt := `You are a helpful assistant. Based on the following expanded context from conversation summaries, answer the user's question.

Provide a focused answer based on the context above.`

	msgs := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("Context:\n%s\n\nQuestion: %s", ctxContent, question)},
	}

	ctx := context.Background()
	result, err := s.client.Complete(ctx, msgs, maxTokens)
	if err != nil {
		s.log.Error(fmt.Sprintf("Expand query failed: %v", err))
		return "", 0, err
	}

	usedTokens := len(result) / 4
	s.log.Info(fmt.Sprintf("Expand query completed, answer length: %d chars", len(result)))
	return result, usedTokens, nil
}
