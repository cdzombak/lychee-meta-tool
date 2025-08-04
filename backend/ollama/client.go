package ollama

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cdzombak/lychee-meta-tool/backend/constants"
	"github.com/ollama/ollama/api"
)

// Constants for configuration and limits (using shared constants)
const (
	// DefaultTimeout for HTTP requests
	DefaultTimeout = constants.OllamaClientTimeout
	// GenerationTimeout for AI title generation
	GenerationTimeout = constants.AIGenerationTimeout
	// MaxImageSize maximum image size to process
	MaxImageSize = constants.MaxImageSize
	// HTTPTimeout for image downloads
	HTTPTimeout = constants.ImageDownloadTimeout
)

// Default prompts for different scenarios
const (
	DetailedPrompt = "Provide a title for this photo. The title should be eloquent and concise, suitable for an artistic photograph but not pretentious. The title should be just a few words at most; shorter is usually better. You MUST provide _only_ the title as your response."
	SimplePrompt   = "Describe this image with a short, artistic title (3-5 words maximum):"
)

// Image format validation (using shared constants where possible)
var (
	validImageTypes = []string{constants.MimeJPEG, "image/jpg", constants.MimePNG, constants.MimeWEBP, constants.MimeGIF}
	validImageSigs  = map[string][]byte{
		"jpeg": {0xFF, 0xD8, 0xFF},
		"png":  {0x89, 0x50, 0x4E, 0x47},
		"gif":  {0x47, 0x49, 0x46, 0x38},
		"webp": {0x52, 0x49, 0x46, 0x46}, // RIFF header, WEBP at offset 8
	}
)

// GenerationStrategy represents different approaches to send images to Ollama
type GenerationStrategy int

const (
	StrategyRawBytes GenerationStrategy = iota
	StrategyBase64
	StrategyDataURI
	StrategyTempFile
)

// Client wraps the Ollama API client with additional functionality
type Client struct {
	client *api.Client
	model  string
}

// NewClient creates a new Ollama client with the specified URL and model
func NewClient(url, model string) (*Client, error) {
	if model == "" {
		return nil, fmt.Errorf("model name is required")
	}

	httpClient := &http.Client{
		Timeout: DefaultTimeout,
	}

	var client *api.Client
	if url != "" {
		parsedURL, err := parseURL(url)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Ollama URL %q: %w", url, err)
		}
		client = api.NewClient(parsedURL, httpClient)
		log.Printf("Ollama client configured with URL: %s, Model: %s", url, model)
	} else {
		var err error
		client, err = api.ClientFromEnvironment()
		if err != nil {
			return nil, fmt.Errorf("failed to create Ollama client from environment: %w", err)
		}
		log.Printf("Ollama client configured from environment, Model: %s", model)
	}

	return &Client{
		client: client,
		model:  model,
	}, nil
}

// parseURL validates and parses a URL string
func parseURL(rawURL string) (*url.URL, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	if u.Scheme == "" {
		return nil, fmt.Errorf("URL must include scheme (http/https)")
	}

	if u.Host == "" {
		return nil, fmt.Errorf("URL must include host")
	}

	return u, nil
}

// GenerateTitle downloads an image and generates a title using Ollama AI
func (c *Client) GenerateTitle(ctx context.Context, imageURL string) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("image URL cannot be empty")
	}

	// Download image with validation
	imageData, contentType, err := c.downloadImage(ctx, imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}

	// Generate title using strategy pattern with fallbacks
	return c.generateTitleWithFallback(ctx, imageData, contentType)
}

// downloadImage downloads and validates an image from the given URL
func (c *Client) downloadImage(ctx context.Context, imageURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	client := http.Client{Timeout: HTTPTimeout}
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

	if !isValidImageType(contentType) {
		return nil, "", fmt.Errorf("unsupported image type: %s", contentType)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	if len(imageData) == 0 {
		return nil, "", fmt.Errorf("received empty image data")
	}

	if !hasValidImageSignature(imageData) {
		return nil, "", fmt.Errorf("invalid image format or corrupted data")
	}

	if len(imageData) > MaxImageSize {
		log.Printf("Warning: Large image detected (%d bytes), may cause performance issues", len(imageData))
	}

	log.Printf("Image validation successful: %d bytes", len(imageData))
	return imageData, contentType, nil
}

// isValidImageType checks if the content type is supported
func isValidImageType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	for _, validType := range validImageTypes {
		if strings.Contains(contentType, validType) {
			return true
		}
	}
	return false
}

// hasValidImageSignature checks if the data starts with a valid image file signature
func hasValidImageSignature(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// Check JPEG signature
	if bytes := validImageSigs["jpeg"]; len(data) >= len(bytes) {
		if compareBytes(data[:len(bytes)], bytes) {
			return true
		}
	}

	// Check PNG signature
	if bytes := validImageSigs["png"]; len(data) >= len(bytes) {
		if compareBytes(data[:len(bytes)], bytes) {
			return true
		}
	}

	// Check GIF signature
	if bytes := validImageSigs["gif"]; len(data) >= len(bytes) {
		if compareBytes(data[:len(bytes)], bytes) {
			return true
		}
	}

	// Check WebP signature (RIFF header + WEBP at offset 8)
	if bytes := validImageSigs["webp"]; len(data) >= len(bytes) {
		if compareBytes(data[:len(bytes)], bytes) && len(data) >= 12 {
			webpSig := []byte{0x57, 0x45, 0x42, 0x50} // "WEBP"
			if compareBytes(data[8:12], webpSig) {
				return true
			}
		}
	}

	return false
}

// compareBytes compares two byte slices for equality
func compareBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// generateTitleWithFallback tries multiple strategies to generate a title
func (c *Client) generateTitleWithFallback(ctx context.Context, imageBytes []byte, contentType string) (string, error) {
	strategies := []GenerationStrategy{
		StrategyRawBytes,
		StrategyBase64,
		StrategyDataURI,
		StrategyTempFile,
	}

	var lastErr error
	for i, strategy := range strategies {
		log.Printf("Attempting strategy %d/%d: %s", i+1, len(strategies), strategyName(strategy))
		
		title, err := c.generateTitleWithStrategy(ctx, imageBytes, contentType, strategy)
		if err == nil && title != "" {
			log.Printf("Success with strategy: %s, title: %s", strategyName(strategy), title)
			return title, nil
		}
		
		lastErr = err
		if err != nil {
			log.Printf("Strategy %s failed: %v", strategyName(strategy), err)
		} else {
			log.Printf("Strategy %s returned empty response", strategyName(strategy))
		}
	}

	return "", fmt.Errorf("all generation strategies failed, last error: %w", lastErr)
}

// strategyName returns a human-readable name for the strategy
func strategyName(strategy GenerationStrategy) string {
	switch strategy {
	case StrategyRawBytes:
		return "raw bytes"
	case StrategyBase64:
		return "base64"
	case StrategyDataURI:
		return "data URI"
	case StrategyTempFile:
		return "temporary file"
	default:
		return "unknown"
	}
}

// generateTitleWithStrategy generates a title using the specified strategy
func (c *Client) generateTitleWithStrategy(ctx context.Context, imageBytes []byte, contentType string, strategy GenerationStrategy) (string, error) {
	var imageData api.ImageData
	var cleanup func()

	switch strategy {
	case StrategyRawBytes:
		imageData = api.ImageData(imageBytes)
		cleanup = func() {} // No cleanup needed

	case StrategyBase64:
		imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)
		imageData = api.ImageData(imageBase64)
		cleanup = func() {} // No cleanup needed

	case StrategyDataURI:
		imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)
		dataURI := fmt.Sprintf("data:%s;base64,%s", contentType, imageBase64)
		imageData = api.ImageData(dataURI)
		cleanup = func() {} // No cleanup needed

	case StrategyTempFile:
		tmpFile, err := c.createTempFile(imageBytes, contentType)
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		imageData = api.ImageData(tmpFile)
		cleanup = func() { os.Remove(tmpFile) }

	default:
		return "", fmt.Errorf("unsupported strategy: %d", strategy)
	}

	defer cleanup()

	// Use detailed prompt for first strategy, simple for fallbacks
	prompt := DetailedPrompt
	if strategy != StrategyRawBytes {
		prompt = SimplePrompt
	}

	return c.executeGeneration(ctx, imageData, prompt)
}

// createTempFile creates a temporary file with the image data
func (c *Client) createTempFile(imageBytes []byte, contentType string) (string, error) {
	ext := getFileExtension(contentType)
	tmpFile, err := os.CreateTemp("", "ollama_image_*"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(imageBytes); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return tmpFile.Name(), nil
}

// executeGeneration performs the actual API call to Ollama
func (c *Client) executeGeneration(ctx context.Context, imageData api.ImageData, prompt string) (string, error) {
	req := &api.GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Images: []api.ImageData{imageData},
		Stream: &[]bool{false}[0],
		Options: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.9,
		},
	}

	var fullResponse strings.Builder
	err := c.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if resp.Response != "" {
			fullResponse.WriteString(resp.Response)
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	result := strings.TrimSpace(fullResponse.String())
	if result == "" {
		return "", fmt.Errorf("received empty response")
	}

	// Clean up the title (remove quotes, trim whitespace)
	result = strings.Trim(result, `"'`)
	return result, nil
}

// getFileExtension returns the appropriate file extension for the content type
func getFileExtension(contentType string) string {
	switch strings.ToLower(contentType) {
	case constants.MimeJPEG, "image/jpg":
		return constants.ExtJPG
	case constants.MimePNG:
		return constants.ExtPNG
	case constants.MimeGIF:
		return constants.ExtGIF
	case constants.MimeWEBP:
		return constants.ExtWEBP
	default:
		return constants.ExtJPG // Default to JPEG
	}
}

