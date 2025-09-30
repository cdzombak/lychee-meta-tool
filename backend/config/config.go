package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cdzombak/lychee-meta-tool/backend/constants"
	"gopkg.in/yaml.v3"
)

// Constants for validation
const (
	// Port ranges
	MinPort = 1
	MaxPort = 65535

	// Default values (using shared constants)
	DefaultServerPort = constants.DefaultServerPort
	DefaultMySQLPort  = constants.DefaultDatabasePort
	DefaultPostgresPort = constants.DefaultPostgresPort

	// Database types
	DatabaseMySQL    = "mysql"
	DatabasePostgres = "postgres"
	DatabaseSQLite   = "sqlite"
)

var modelNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._:/\-]+$`)

type DatabaseConfig struct {
	Type     string `yaml:"type" json:"type"`
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Database string `yaml:"database" json:"database"`
	Path     string `yaml:"path" json:"path"` // For SQLite
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
}

type ServerConfig struct {
	Port int        `yaml:"port" json:"port"`
	CORS CORSConfig `yaml:"cors" json:"cors"`
}

type OllamaConfig struct {
	URL   string `yaml:"url" json:"url"`
	Model string `yaml:"model" json:"model"`
}

type OpenAIConfig struct {
	URL    string `yaml:"url" json:"url"`
	APIKey string `yaml:"api_key" json:"api_key"`
	Model  string `yaml:"model" json:"model"`
}

type Config struct {
	Database      DatabaseConfig `yaml:"database" json:"database"`
	Server        ServerConfig   `yaml:"server" json:"server"`
	LycheeBaseURL string         `yaml:"lychee_base_url" json:"lychee_base_url"`
	Ollama        OllamaConfig   `yaml:"ollama" json:"ollama"`
	OpenAI        OpenAIConfig   `yaml:"openai" json:"openai"`
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	ext := filepath.Ext(configPath)

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validate performs comprehensive validation of the configuration
func (c *Config) validate() error {
	// Apply defaults first
	c.applyDefaults()

	// Validate database configuration
	if err := c.validateDatabase(); err != nil {
		return fmt.Errorf("database configuration error: %w", err)
	}

	// Validate server configuration
	if err := c.validateServer(); err != nil {
		return fmt.Errorf("server configuration error: %w", err)
	}

	// Validate Lychee base URL
	if err := c.validateLycheeBaseURL(); err != nil {
		return fmt.Errorf("lychee_base_url configuration error: %w", err)
	}

	// Validate Ollama configuration (optional)
	if err := c.validateOllama(); err != nil {
		return fmt.Errorf("ollama configuration error: %w", err)
	}

	// Validate OpenAI configuration (optional)
	if err := c.validateOpenAI(); err != nil {
		return fmt.Errorf("openai configuration error: %w", err)
	}

	// Ensure only one AI backend is configured
	if err := c.validateAIBackendExclusivity(); err != nil {
		return fmt.Errorf("AI backend configuration error: %w", err)
	}

	return nil
}

// applyDefaults sets default values for optional configuration fields
func (c *Config) applyDefaults() {
	// Set default server port
	if c.Server.Port == 0 {
		c.Server.Port = DefaultServerPort
	}

	// Set default database ports
	if c.Database.Port == 0 {
		switch c.Database.Type {
		case DatabaseMySQL:
			c.Database.Port = DefaultMySQLPort
		case DatabasePostgres:
			c.Database.Port = DefaultPostgresPort
		}
	}

	// Ensure CORS origins is not nil
	if c.Server.CORS.AllowedOrigins == nil {
		c.Server.CORS.AllowedOrigins = []string{}
	}
}

// validateDatabase validates database configuration
func (c *Config) validateDatabase() error {
	if c.Database.Type == "" {
		return fmt.Errorf("type is required (supported: mysql, postgres, sqlite)")
	}

	switch c.Database.Type {
	case DatabaseMySQL, DatabasePostgres:
		if c.Database.Host == "" {
			return fmt.Errorf("host is required for %s database", c.Database.Type)
		}
		if c.Database.User == "" {
			return fmt.Errorf("user is required for %s database", c.Database.Type)
		}
		if c.Database.Database == "" {
			return fmt.Errorf("database name is required for %s database", c.Database.Type)
		}
		if c.Database.Port < MinPort || c.Database.Port > MaxPort {
			return fmt.Errorf("port must be between %d and %d, got %d", MinPort, MaxPort, c.Database.Port)
		}

	case DatabaseSQLite:
		if c.Database.Path == "" {
			return fmt.Errorf("path is required for sqlite database")
		}
		// Validate that the directory exists (create if parent doesn't exist is common pattern)
		dir := filepath.Dir(c.Database.Path)
		if dir != "." && dir != "" {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return fmt.Errorf("directory for sqlite database does not exist: %s", dir)
			}
		}

	default:
		return fmt.Errorf("unsupported database type: %s (supported: mysql, postgres, sqlite)", c.Database.Type)
	}

	return nil
}

// validateServer validates server configuration
func (c *Config) validateServer() error {
	if c.Server.Port < MinPort || c.Server.Port > MaxPort {
		return fmt.Errorf("port must be between %d and %d, got %d", MinPort, MaxPort, c.Server.Port)
	}

	// Validate CORS origins
	for i, origin := range c.Server.CORS.AllowedOrigins {
		if origin == "" {
			return fmt.Errorf("CORS allowed_origins[%d] cannot be empty", i)
		}
		// Parse URL to validate format
		if _, err := url.Parse(origin); err != nil {
			return fmt.Errorf("CORS allowed_origins[%d] has invalid URL format %q: %w", i, origin, err)
		}
	}

	return nil
}

// validateLycheeBaseURL validates the Lychee base URL
func (c *Config) validateLycheeBaseURL() error {
	if c.LycheeBaseURL == "" {
		return fmt.Errorf("lychee_base_url is required")
	}

	parsedURL, err := url.Parse(c.LycheeBaseURL)
	if err != nil {
		return fmt.Errorf("invalid URL format %q: %w", c.LycheeBaseURL, err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("lychee_base_url must include scheme (http/https): %q", c.LycheeBaseURL)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("lychee_base_url must include host: %q", c.LycheeBaseURL)
	}

	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		return fmt.Errorf("lychee_base_url must use http or https scheme, got: %q", parsedURL.Scheme)
	}

	return nil
}

// validateOllama validates Ollama configuration (optional)
func (c *Config) validateOllama() error {
	// Ollama configuration is optional - if URL is empty, skip validation
	if c.Ollama.URL == "" && c.Ollama.Model == "" {
		return nil // No Ollama configuration provided
	}

	// If one field is provided, both should be provided
	if c.Ollama.URL == "" {
		return fmt.Errorf("url is required when model is specified")
	}
	if c.Ollama.Model == "" {
		return fmt.Errorf("model is required when url is specified")
	}

	// Validate URL format
	parsedURL, err := url.Parse(c.Ollama.URL)
	if err != nil {
		return fmt.Errorf("invalid URL format %q: %w", c.Ollama.URL, err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("url must include scheme (http/https): %q", c.Ollama.URL)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("url must include host: %q", c.Ollama.URL)
	}

	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		return fmt.Errorf("url must use http or https scheme, got: %q", parsedURL.Scheme)
	}

	if !modelNamePattern.MatchString(c.Ollama.Model) {
		return fmt.Errorf("model name contains invalid characters (allowed: alphanumeric, dots, colons, hyphens, slashes): %q", c.Ollama.Model)
	}

	return nil
}

// GetDSN returns the database connection string for the configured database
func (c *Config) GetDSN() string {
	switch c.Database.Type {
	case DatabaseMySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
			c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Database)
	case DatabasePostgres:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.Database)
	case DatabaseSQLite:
		return c.Database.Path
	default:
		return ""
	}
}

// validateOpenAI validates OpenAI configuration (optional)
func (c *Config) validateOpenAI() error {
	// OpenAI configuration is optional - if URL is empty, skip validation
	if c.OpenAI.URL == "" && c.OpenAI.APIKey == "" && c.OpenAI.Model == "" {
		return nil
	}

	// If one field is provided, URL and APIKey are required
	if c.OpenAI.URL == "" {
		return fmt.Errorf("url is required when openai is configured")
	}
	if c.OpenAI.APIKey == "" {
		return fmt.Errorf("api_key is required when openai is configured")
	}

	// Validate URL format
	parsedURL, err := url.Parse(c.OpenAI.URL)
	if err != nil {
		return fmt.Errorf("invalid URL format %q: %w", c.OpenAI.URL, err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("url must include scheme (http/https): %q", c.OpenAI.URL)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("url must include host: %q", c.OpenAI.URL)
	}

	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		return fmt.Errorf("url must use http or https scheme, got: %q", parsedURL.Scheme)
	}

	if c.OpenAI.Model != "" && !modelNamePattern.MatchString(c.OpenAI.Model) {
		return fmt.Errorf("model name contains invalid characters (allowed: alphanumeric, dots, colons, hyphens, slashes): %q", c.OpenAI.Model)
	}

	return nil
}

// validateAIBackendExclusivity ensures only one AI backend is configured
func (c *Config) validateAIBackendExclusivity() error {
	ollamaEnabled := c.Ollama.URL != "" && c.Ollama.Model != ""
	openAIEnabled := c.OpenAI.URL != "" && c.OpenAI.APIKey != ""

	if ollamaEnabled && openAIEnabled {
		return fmt.Errorf("cannot configure both Ollama and OpenAI backends simultaneously. Please choose one")
	}

	return nil
}

// IsOllamaEnabled returns true if Ollama configuration is provided and valid
func (c *Config) IsOllamaEnabled() bool {
	return c.Ollama.URL != "" && c.Ollama.Model != ""
}

// IsOpenAIEnabled returns true if OpenAI configuration is provided and valid
func (c *Config) IsOpenAIEnabled() bool {
	return c.OpenAI.URL != "" && c.OpenAI.APIKey != ""
}
