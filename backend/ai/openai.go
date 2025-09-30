package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/cdzombak/lychee-meta-tool/backend/constants"
)

const (
	DefaultModel = "gpt-4o"
	SystemPrompt = "You are a professional photo curator. Provide concise, eloquent titles for artistic photographs. The title should be just a few words, never more than 10 words. You MUST provide only the title as your response, nothing else."
	UserPrompt   = "Provide a title for this photograph. The title should be eloquent and concise, suitable for an artistic photograph but not pretentious. The title should be just a few words at most; shorter is usually better. You MUST provide _only_ the title as your response."
)

type OpenAIClient struct {
	apiURL string
	apiKey string
	model  string
	client *http.Client
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	MaxTokens int            `json:"max_tokens"`
}

type openAIMessage struct {
	Role    string                   `json:"role"`
	Content []openAIMessageContent   `json:"content"`
}

type openAIMessageContent struct {
	Type     string                  `json:"type"`
	Text     string                  `json:"text,omitempty"`
	ImageURL *openAIImageURL         `json:"image_url,omitempty"`
}

type openAIImageURL struct {
	URL    string `json:"url"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func NewOpenAIClient(apiURL, apiKey, model string) (*OpenAIClient, error) {
	if apiURL == "" {
		return nil, fmt.Errorf("API URL is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if model == "" {
		model = DefaultModel
	}

	client := &http.Client{
		Timeout: constants.OllamaClientTimeout,
	}

	log.Printf("OpenAI client configured with URL: %s, Model: %s", apiURL, model)

	return &OpenAIClient{
		apiURL: apiURL,
		apiKey: apiKey,
		model:  model,
		client: client,
	}, nil
}

func (c *OpenAIClient) GenerateTitle(ctx context.Context, imageURL string) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("image URL cannot be empty")
	}

	imageData, contentType, err := downloadImage(ctx, imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	dataURI := fmt.Sprintf("data:%s;base64,%s", contentType, base64Image)

	reqBody := openAIRequest{
		Model: c.model,
		Messages: []openAIMessage{
			{
				Role: "system",
				Content: []openAIMessageContent{
					{
						Type: "text",
						Text: SystemPrompt,
					},
				},
			},
			{
				Role: "user",
				Content: []openAIMessageContent{
					{
						Type: "text",
						Text: UserPrompt,
					},
					{
						Type: "image_url",
						ImageURL: &openAIImageURL{
							URL: dataURI,
						},
					},
				},
			},
		},
		MaxTokens: 50,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	log.Printf("Sending request to OpenAI-style endpoint for image: %s", imageURL)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenAI API error (HTTP %d): %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp openAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("API error: %s (%s)", apiResp.Error.Message, apiResp.Error.Type)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	title := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	title = strings.Trim(title, `"'`)

	if title == "" {
		return "", fmt.Errorf("received empty title")
	}

	log.Printf("Successfully generated title: %s", title)
	return title, nil
}

func downloadImage(ctx context.Context, imageURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	client := http.Client{Timeout: constants.ImageDownloadTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	log.Printf("Downloaded image: Content-Type=%s, Status=%d, URL=%s", contentType, resp.StatusCode, imageURL)

	validTypes := []string{constants.MimeJPEG, "image/jpg", constants.MimePNG, constants.MimeWEBP, constants.MimeGIF}
	valid := false
	contentTypeLower := strings.ToLower(contentType)
	for _, validType := range validTypes {
		if strings.Contains(contentTypeLower, validType) {
			valid = true
			break
		}
	}
	if !valid {
		return nil, "", fmt.Errorf("unsupported image type: %s", contentType)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	if len(imageData) == 0 {
		return nil, "", fmt.Errorf("received empty image data")
	}

	if len(imageData) > constants.MaxImageSize {
		log.Printf("Warning: Large image detected (%d bytes), may cause performance issues", len(imageData))
	}

	log.Printf("Image validation successful: %d bytes", len(imageData))
	return imageData, contentType, nil
}
