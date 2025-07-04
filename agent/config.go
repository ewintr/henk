package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"go-mod.ewintr.nl/henk/agent/llm"
)

type Config struct {
	DefaultProvider string         `toml:"default_provider"`
	DefaultModel    string         `toml:"default_model"`
	Providers       []llm.Provider `toml:"providers"`
	SystemPrompt    string         `toml:"system_prompt"`
}

func (c Config) Validate() error {
	if len(c.Providers) == 0 {
		return fmt.Errorf("no providers configured")
	}

	// Validate all providers have models
	for i, provider := range c.Providers {
		if len(provider.Models) == 0 {
			return fmt.Errorf("provider %d has no models configured", i)
		}
	}

	// Find default models and validate
	var defaultCount int
	for _, provider := range c.Providers {
		for _, model := range provider.Models {
			if model.Default {
				defaultCount++
			}
		}
	}

	if defaultCount > 1 {
		return fmt.Errorf("multiple models configured as default")
	}

	return nil
}

func (c Config) Provider(name string) (llm.Provider, bool) {
	for _, provider := range c.Providers {
		if provider.Name == name {
			return provider, true
		}
	}

	return llm.Provider{}, false
}

func (c Config) ProviderByModelName(name string) (llm.Provider, bool) {
	for _, provider := range c.Providers {
		if _, ok := provider.Model(name); ok {
			return provider, true
		}
	}

	return llm.Provider{}, false
}

func ReadConfig() (Config, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, fmt.Errorf("could not find user config dir: %v", err)
	}
	configDir := filepath.Join(userConfigDir, "henk")
	if err := setupDir(configDir); err != nil {
		return Config{}, fmt.Errorf("could not create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")

	var config Config
	_, err = toml.DecodeFile(configPath, &config)
	if err != nil {
		return Config{}, fmt.Errorf("could not read config file: %v", err)
	}

	// set keys
	for i, p := range config.Providers {
		if p.ApiKeyEnv != "" {
			val, ok := os.LookupEnv(p.ApiKeyEnv)
			if !ok {
				return Config{}, fmt.Errorf("could not read environment variable %s", p.ApiKeyEnv)
			}
			p.ApiKey = val
		}
		config.Providers[i] = p
	}

	// default values
	if config.SystemPrompt == "" {
		config.SystemPrompt = "You are a helpful assistent. Be concise and accurate in your responses."
	}

	return config, nil
}

func setupDir(path string) error {
	info, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("%s exists, but is not a directory", path)
	}

	return nil
}
