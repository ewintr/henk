package main

import (
	"fmt"
	"go-mod.ewintr.nl/henk/llm"
)

type Config struct {
	Providers []llm.Provider `toml:"providers"`
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

func (c Config) Provider() llm.Provider {
	// Find provider with default model
	for _, provider := range c.Providers {
		for _, model := range provider.Models {
			if model.Default {
				return provider
			}
		}
	}

	// If no default, return first provider
	if len(c.Providers) > 0 {
		return c.Providers[0]
	}

	// Fallback (should not happen if Validate() passed)
	return llm.Provider{}
}
