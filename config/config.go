package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	UserName                 string            `yaml:"userName" json:"userName" toml:"userName"`
	MainBranchName           string            `yaml:"mainBranchName" json:"mainBranchName" toml:"mainBranchName"`
	DocBranchNameTemplate    string            `yaml:"docBranchNameTemplate" json:"docBranchNameTemplate" toml:"docBranchNameTemplate"`
	DocWorktreeDir           string            `yaml:"docWorktreeDir" json:"docWorktreeDir" toml:"docWorktreeDir"`
	AIAgentMemoryContextPath map[string]string `yaml:"aIAgentMemoryContextPath" json:"aIAgentMemoryContextPath" toml:"aIAgentMemoryContextPath"`
	IgnorePatterns           []string          `yaml:"ignorePatterns" json:"ignorePatterns" toml:"ignorePatterns"`
	DocDir                   string            `yaml:"docDir" json:"docDir" toml:"docDir"`
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = ".ai-docs.config.yml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{
		MainBranchName:        "main",
		DocBranchNameTemplate: "@doc/{userName}",
		DocWorktreeDir:        ".mem",
		AIAgentMemoryContextPath: map[string]string{
			"Cline":  "memory-bank",
			"Claude": ".ai-memory",
			"Gemini": ".gemini/context",
			"Cursor": ".cursor/rules",
		},
		IgnorePatterns: []string{
			"/memory-bank/",
			"/.ai-memory/",
			"/.gemini/context/",
			"/.cursor/rules/",
		},
		DocDir: "docs/ai",
	}

	ext := filepath.Ext(configPath)
	switch ext {
	case ".yml", ".yaml":
		err = yaml.Unmarshal(data, cfg)
	case ".json":
		err = json.Unmarshal(data, cfg)
	case ".toml":
		err = toml.Unmarshal(data, cfg)
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.UserName == "" {
		cfg.UserName = getGitUserName()
	}

	return cfg, nil
}

func getGitUserName() string {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		return strings.TrimSpace(string(output))
	}

	whoamiCmd := exec.Command("whoami")
	whoamiOutput, err := whoamiCmd.Output()
	if err != nil || strings.TrimSpace(string(whoamiOutput)) == "" {
		return "user"
	}
	return strings.TrimSpace(string(whoamiOutput))
}

func (c *Config) GetDocBranchName() string {
	return strings.ReplaceAll(c.DocBranchNameTemplate, "{userName}", c.UserName)
}
