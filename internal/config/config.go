package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Linear   LinearConfig   `mapstructure:"linear"`
	Defaults DefaultsConfig `mapstructure:"defaults"`
	UI       UIConfig       `mapstructure:"ui"`
	Git      GitConfig      `mapstructure:"git"`
	AI       AIConfig       `mapstructure:"ai"`
}

// LinearConfig holds Linear API configuration
type LinearConfig struct {
	APIKey string `mapstructure:"api_key"`
}

// DefaultsConfig holds default view settings
type DefaultsConfig struct {
	Team    string `mapstructure:"team"`
	Project string `mapstructure:"project"`
	View    string `mapstructure:"view"`
}

// UIConfig holds UI preferences
type UIConfig struct {
	Theme      string `mapstructure:"theme"`
	VimMode    bool   `mapstructure:"vim_mode"`
	ShowIDs    bool   `mapstructure:"show_ids"`
	DateFormat string `mapstructure:"date_format"`
}

// GitConfig holds git integration settings
type GitConfig struct {
	BranchPrefix string `mapstructure:"branch_prefix"`
	BranchFormat string `mapstructure:"branch_format"`
}

// AIConfig holds AI provider configuration
type AIConfig struct {
	Provider  string          `mapstructure:"provider"`
	OpenAI    OpenAIConfig    `mapstructure:"openai"`
	Anthropic AnthropicConfig `mapstructure:"anthropic"`
	Ollama    OllamaConfig    `mapstructure:"ollama"`
}

// OpenAIConfig holds OpenAI settings
type OpenAIConfig struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}

// AnthropicConfig holds Anthropic settings
type AnthropicConfig struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}

// OllamaConfig holds Ollama settings
type OllamaConfig struct {
	Host  string `mapstructure:"host"`
	Model string `mapstructure:"model"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add config paths
	if configDir, err := os.UserConfigDir(); err == nil {
		v.AddConfigPath(filepath.Join(configDir, "lazyliner"))
	}
	v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "lazyliner"))
	v.AddConfigPath(".")

	// Set defaults
	setDefaults(v)

	// Environment variable support
	v.SetEnvPrefix("LAZYLINER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Also check for LINEAR_API_KEY as fallback
	if apiKey := os.Getenv("LAZYLINER_API_KEY"); apiKey != "" {
		v.Set("linear.api_key", apiKey)
	} else if apiKey := os.Getenv("LINEAR_API_KEY"); apiKey != "" {
		v.Set("linear.api_key", apiKey)
	}

	// Read config file (ignore if not found)
	_ = v.ReadInConfig()

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Linear defaults
	v.SetDefault("linear.api_key", "")

	// Defaults
	v.SetDefault("defaults.team", "")
	v.SetDefault("defaults.project", "")
	v.SetDefault("defaults.view", "my-issues")

	// UI defaults
	v.SetDefault("ui.theme", "dark")
	v.SetDefault("ui.vim_mode", true)
	v.SetDefault("ui.show_ids", true)
	v.SetDefault("ui.date_format", "relative")

	// Git defaults
	v.SetDefault("git.branch_prefix", "feature")
	v.SetDefault("git.branch_format", "{prefix}/{id}-{title}")

	// AI defaults
	v.SetDefault("ai.provider", "openai")
	v.SetDefault("ai.openai.api_key", "")
	v.SetDefault("ai.openai.model", "gpt-4")
	v.SetDefault("ai.anthropic.api_key", "")
	v.SetDefault("ai.anthropic.model", "claude-3-sonnet-20240229")
	v.SetDefault("ai.ollama.host", "http://localhost:11434")
	v.SetDefault("ai.ollama.model", "llama2")
}

// ConfigDir returns the configuration directory path
func ConfigDir() string {
	if configDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(configDir, "lazyliner")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "lazyliner")
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0755)
}
