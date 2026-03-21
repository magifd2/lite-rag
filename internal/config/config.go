// Package config loads and validates application configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the top-level configuration structure.
type Config struct {
	API       APIConfig       `toml:"api"`
	Models    ModelsConfig    `toml:"models"`
	Database  DatabaseConfig  `toml:"database"`
	Retrieval RetrievalConfig `toml:"retrieval"`
	Server    ServerConfig    `toml:"server"`
}

// ServerConfig controls the HTTP server started by the serve subcommand.
type ServerConfig struct {
	Addr     string `toml:"addr"`
	LogLevel string `toml:"log_level"`
}

// APIConfig holds LLM API connection settings.
type APIConfig struct {
	BaseURL string `toml:"base_url"`
	APIKey  string `toml:"api_key"`
}

// ModelsConfig specifies which models to use.
type ModelsConfig struct {
	Embedding string `toml:"embedding"`
	Chat      string `toml:"chat"`
}

// DatabaseConfig specifies the DuckDB file path.
type DatabaseConfig struct {
	Path string `toml:"path"`
}

// RetrievalConfig controls chunking and retrieval behaviour.
type RetrievalConfig struct {
	TopK          int  `toml:"top_k"`
	ContextWindow int  `toml:"context_window"`
	ChunkSize     int  `toml:"chunk_size"`
	ChunkOverlap  int  `toml:"chunk_overlap"`
	QueryRewrite  bool `toml:"query_rewrite"`
}

// defaults returns a Config populated with sensible defaults.
func defaults() Config {
	return Config{
		API: APIConfig{
			BaseURL: "http://localhost:1234/v1",
			APIKey:  "lm-studio",
		},
		Models: ModelsConfig{
			Embedding: "nomic-ai/nomic-embed-text-v1.5-GGUF",
			Chat:      "openai/gpt-oss-20b",
		},
		Database: DatabaseConfig{
			Path: "./lite-rag.db",
		},
		Retrieval: RetrievalConfig{
			TopK:          5,
			ContextWindow: 1,
			ChunkSize:     512,
			ChunkOverlap:  64,
		},
		Server: ServerConfig{
			Addr:     "127.0.0.1:8080",
			LogLevel: "info",
		},
	}
}

// DefaultConfigPath returns the XDG-compliant default config file path:
// $XDG_CONFIG_HOME/lite-rag/config.toml, or ~/.config/lite-rag/config.toml.
func DefaultConfigPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		if home, err := os.UserHomeDir(); err == nil {
			base = filepath.Join(home, ".config")
		}
	}
	if base != "" {
		return filepath.Join(base, "lite-rag", "config.toml")
	}
	return "config.toml"
}

// Load reads configuration from the given file path and overlays environment
// variable overrides. The file need not exist; missing files yield defaults.
func Load(path string) (*Config, error) {
	cfg := defaults()

	if _, err := os.Stat(path); err == nil {
		if _, err := toml.DecodeFile(path, &cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", path, err)
		}
	}

	applyEnv(&cfg)
	return &cfg, nil
}

// applyEnv overlays LITE_RAG_* environment variables onto cfg.
func applyEnv(cfg *Config) {
	if v := os.Getenv("LITE_RAG_API_BASE_URL"); v != "" {
		cfg.API.BaseURL = v
	}
	if v := os.Getenv("LITE_RAG_API_KEY"); v != "" {
		cfg.API.APIKey = v
	}
	if v := os.Getenv("LITE_RAG_EMBEDDING_MODEL"); v != "" {
		cfg.Models.Embedding = v
	}
	if v := os.Getenv("LITE_RAG_CHAT_MODEL"); v != "" {
		cfg.Models.Chat = v
	}
	if v := os.Getenv("LITE_RAG_DB_PATH"); v != "" {
		cfg.Database.Path = v
	}
}
