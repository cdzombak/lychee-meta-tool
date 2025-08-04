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
	"time"

	"github.com/ollama/ollama/api"
)

type Client struct {
	client *api.Client
	model  string
}

func NewClient(url, model string) (*Client, error) {
	httpClient := &http.Client{
		Timeout: 5 * time.Minute,
	}

	var client *api.Client
	if url != "" {
		// Parse the URL and create client
		parsedURL, err := parseURL(url)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Ollama URL: %w", err)
		}
		client = api.NewClient(parsedURL, httpClient)
	} else {
		var err error
		client, err = api.ClientFromEnvironment()
		if err != nil {
			return nil, fmt.Errorf("failed to create Ollama client: %w", err)
		}
	}

	return &Client{
		client: client,
		model:  model,
	}, nil
}

func parseURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (c *Client) GenerateTitle(ctx context.Context, imageURL string) (string, error) {
	// Download the image with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	log.Printf("Image download - Content-Type: %s, Status: %d, URL: %s", contentType, resp.StatusCode, imageURL)
	
	if !isValidImageType(contentType) {
		return "", fmt.Errorf("unsupported image type: %s", contentType)
	}

	// Read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Validate image data is not empty
	if len(imageData) == 0 {
		return "", fmt.Errorf("received empty image data")
	}

	log.Printf("Image data loaded successfully: %d bytes", len(imageData))

	// Check if the image data starts with valid image signatures
	if !hasValidImageSignature(imageData) {
		return "", fmt.Errorf("image data does not have valid image signature")
	}

	// If image is very large, we might need to reduce it or use different approach
	if len(imageData) > 5*1024*1024 { // 5MB
		log.Printf("Warning: Large image detected (%d bytes), this may cause issues with Ollama", len(imageData))
	}

	// Use raw image bytes directly (not base64) - this is the key fix based on working reference
	log.Printf("Using raw image bytes directly with Ollama")

	// Try the working approach first
	return c.generateTitleWithRawBytes(ctx, imageData)
}

func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg", 
		"image/png",
		"image/webp",
		"image/gif",
	}
	
	contentType = strings.ToLower(contentType)
	for _, validType := range validTypes {
		if strings.Contains(contentType, validType) {
			return true
		}
	}
	return false
}

func hasValidImageSignature(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// Check for common image file signatures
	// JPEG: FF D8 FF
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return true
	}
	
	// PNG: 89 50 4E 47
	if len(data) >= 4 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return true
	}
	
	// GIF: 47 49 46 38
	if len(data) >= 4 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
		return true
	}
	
	// WebP: 52 49 46 46 (RIFF)
	if len(data) >= 4 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		// Check for WebP signature at offset 8: 57 45 42 50 (WEBP)
		if len(data) >= 12 && data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
			return true
		}
	}
	
	return false
}

func (c *Client) generateTitleWithRawBytes(ctx context.Context, imageBytes []byte) (string, error) {
	log.Printf("Attempting Ollama generation with raw bytes (like working reference), model: %s", c.model)
	
	req := &api.GenerateRequest{
		Model:  c.model,
		Prompt: "Provide a title for this photo. The title should be eloquent and concise, suitable for an artistic photograph but not pretentious. The title should be just a few words at most; shorter is usually better. You MUST provide _only_ the title as your response.",
		Images: []api.ImageData{api.ImageData(imageBytes)},
		Stream: &[]bool{false}[0],
	}

	var fullResponse string
	err := c.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		fullResponse += resp.Response
		if resp.Done {
			log.Printf("Ollama response complete (raw bytes)")
		}
		return nil
	})

	if err != nil {
		log.Printf("Generate with raw bytes failed: %v", err)
		// Fallback to base64 approach if needed
		imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)
		return c.generateTitleSimple(ctx, imageBase64, "image/jpeg")
	}

	result := strings.TrimSpace(fullResponse)
	if result == "" {
		log.Printf("Empty response from raw bytes generate")
		imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)
		return c.generateTitleSimple(ctx, imageBase64, "image/jpeg")
	}

	log.Printf("Successful response from Ollama (raw bytes): %s", result)
	return result, nil
}

func (c *Client) generateTitleSimple(ctx context.Context, imageBase64, contentType string) (string, error) {
	log.Printf("Attempting simple Ollama generation with model: %s", c.model)
	
	// Try raw base64 first
	req := &api.GenerateRequest{
		Model:  c.model,
		Prompt: "Describe this image with a short, artistic title (3-5 words maximum):",
		Images: []api.ImageData{api.ImageData(imageBase64)},
		Stream: &[]bool{false}[0],
	}

	var fullResponse string
	err := c.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		fullResponse += resp.Response
		if resp.Done {
			log.Printf("Ollama response complete")
		}
		return nil
	})

	if err != nil {
		log.Printf("Simple generate with raw base64 failed: %v", err)
		// Try with data URI format
		return c.generateTitleWithDataURI(ctx, imageBase64, contentType)
	}

	result := strings.TrimSpace(fullResponse)
	if result == "" {
		log.Printf("Empty response from simple generate")
		return c.generateTitleWithDataURI(ctx, imageBase64, contentType)
	}

	log.Printf("Successful response from Ollama: %s", result)
	return result, nil
}

func (c *Client) generateTitleWithDataURI(ctx context.Context, imageBase64, contentType string) (string, error) {
	log.Printf("Attempting Ollama generation with data URI format")
	
	// Create data URI: data:image/jpeg;base64,<base64data>
	dataURI := fmt.Sprintf("data:%s;base64,%s", contentType, imageBase64)
	log.Printf("Data URI length: %d characters", len(dataURI))
	
	req := &api.GenerateRequest{
		Model:  c.model,
		Prompt: "Describe this image with a short, artistic title (3-5 words maximum):",
		Images: []api.ImageData{api.ImageData(dataURI)},
		Stream: &[]bool{false}[0],
	}

	var fullResponse string
	err := c.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		fullResponse += resp.Response
		if resp.Done {
			log.Printf("Ollama response complete (data URI)")
		}
		return nil
	})

	if err != nil {
		log.Printf("Generate with data URI failed: %v", err)
		// Try with temporary file approach as last resort
		return c.generateTitleWithTempFile(ctx, imageBase64, contentType)
	}

	result := strings.TrimSpace(fullResponse)
	if result == "" {
		log.Printf("Empty response from data URI generate")
		return c.generateTitleWithTempFile(ctx, imageBase64, contentType)
	}

	log.Printf("Successful response from Ollama (data URI): %s", result)
	return result, nil
}

func (c *Client) generateTitleWithTempFile(ctx context.Context, imageBase64, contentType string) (string, error) {
	log.Printf("Attempting Ollama generation with temporary file approach")
	
	// Decode base64 back to bytes
	imageBytes, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	
	// Determine file extension from content type
	ext := getFileExtension(contentType)
	
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "ollama_image_*"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up
	defer tmpFile.Close()
	
	// Write image data to temp file
	if _, err := tmpFile.Write(imageBytes); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close() // Close before using with Ollama
	
	log.Printf("Created temporary file: %s", tmpFile.Name())
	
	// Try using file path instead of base64
	req := &api.GenerateRequest{
		Model:  c.model,
		Prompt: "Describe this image with a short, artistic title (3-5 words maximum):",
		Images: []api.ImageData{api.ImageData(tmpFile.Name())},
		Stream: &[]bool{false}[0],
	}

	var fullResponse string
	err = c.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		fullResponse += resp.Response
		if resp.Done {
			log.Printf("Ollama response complete (temp file)")
		}
		return nil
	})

	if err != nil {
		log.Printf("Generate with temp file failed: %v", err)
		return c.generateTitleWithGenerate(ctx, imageBase64)
	}

	result := strings.TrimSpace(fullResponse)
	if result == "" {
		log.Printf("Empty response from temp file generate")
		return c.generateTitleWithGenerate(ctx, imageBase64)
	}

	log.Printf("Successful response from Ollama (temp file): %s", result)
	return result, nil
}

func getFileExtension(contentType string) string {
	switch strings.ToLower(contentType) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg" // Default to jpg
	}
}

func (c *Client) generateTitleWithGenerate(ctx context.Context, imageBase64 string) (string, error) {
	// Convert string to ImageData properly
	imageData := api.ImageData(imageBase64)
	log.Printf("Sending request to Ollama Generate endpoint with model: %s, image data length: %d", c.model, len(imageBase64))
	
	req := &api.GenerateRequest{
		Model:  c.model,
		Prompt: "Provide a title for this photo. The title should be eloquent and concise, suitable for an artistic photograph but not pretentious. The title should be just a few words at most; shorter is usually better. You MUST provide _only_ the title as your response.",
		Images: []api.ImageData{imageData},
		Stream: &[]bool{false}[0],
		Options: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.9,
		},
	}

	var response strings.Builder
	err := c.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if resp.Response != "" {
			response.WriteString(resp.Response)
		}
		return nil
	})

	if err != nil {
		log.Printf("Generate endpoint failed: %v, trying Chat endpoint", err)
		return c.generateTitleWithChat(ctx, imageBase64)
	}

	result := strings.TrimSpace(response.String())
	if result == "" {
		log.Printf("Generate endpoint returned empty response, trying Chat endpoint")
		return c.generateTitleWithChat(ctx, imageBase64)
	}

	return result, nil
}

func (c *Client) generateTitleWithChat(ctx context.Context, imageBase64 string) (string, error) {
	// Convert string to ImageData properly  
	imageData := api.ImageData(imageBase64)
	log.Printf("Sending request to Ollama Chat endpoint with model: %s, image data length: %d", c.model, len(imageBase64))
	
	// Prepare the request
	req := &api.ChatRequest{
		Model: c.model,
		Messages: []api.Message{
			{
				Role:    "user",
				Content: "Provide a title for this photo. The title should be eloquent and concise, suitable for an artistic photograph but not pretentious. The title should be just a few words at most; shorter is usually better. You MUST provide _only_ the title as your response.",
				Images:  []api.ImageData{imageData},
			},
		},
		Stream: &[]bool{false}[0],
		Options: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.9,
		},
	}

	var response strings.Builder
	err := c.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		if resp.Message.Content != "" {
			response.WriteString(resp.Message.Content)
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate title with chat: %w", err)
	}

	result := strings.TrimSpace(response.String())
	if result == "" {
		return "", fmt.Errorf("received empty response from Ollama")
	}

	return result, nil
}