package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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

type Config struct {
	Database     DatabaseConfig `yaml:"database" json:"database"`
	Server       ServerConfig   `yaml:"server" json:"server"`
	LycheeBaseURL string        `yaml:"lychee_base_url" json:"lychee_base_url"`
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

func (c *Config) validate() error {
	if c.Database.Type == "" {
		return fmt.Errorf("database type is required")
	}

	switch c.Database.Type {
	case "mysql", "postgres":
		if c.Database.Host == "" {
			return fmt.Errorf("database host is required for %s", c.Database.Type)
		}
		if c.Database.User == "" {
			return fmt.Errorf("database user is required for %s", c.Database.Type)
		}
		if c.Database.Database == "" {
			return fmt.Errorf("database name is required for %s", c.Database.Type)
		}
	case "sqlite":
		if c.Database.Path == "" {
			return fmt.Errorf("database path is required for sqlite")
		}
	default:
		return fmt.Errorf("unsupported database type: %s", c.Database.Type)
	}

	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}

	if c.LycheeBaseURL == "" {
		return fmt.Errorf("lychee_base_url is required")
	}

	return nil
}

func (c *Config) GetDSN() string {
	switch c.Database.Type {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Database)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.Database)
	case "sqlite":
		return c.Database.Path
	default:
		return ""
	}
}