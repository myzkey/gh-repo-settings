package config

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// resolveExtends resolves extends references and merges configurations
func resolveExtends(config *Config, basePath string, visited map[string]bool) (*Config, error) {
	if len(config.Extends) == 0 {
		return config, nil
	}

	// Start with empty base config
	merged := &Config{}

	// Process each extend in order (later ones override earlier ones)
	for _, extendRef := range config.Extends {
		// Normalize the reference for cycle detection
		normalizedRef := normalizeRef(extendRef, basePath)
		if visited[normalizedRef] {
			return nil, fmt.Errorf("circular reference detected: %s", extendRef)
		}
		visited[normalizedRef] = true

		// Load the extended config
		extConfig, newBasePath, err := loadExtendedConfig(extendRef, basePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load extended config %s: %w", extendRef, err)
		}

		// Recursively resolve extends in the loaded config
		if len(extConfig.Extends) > 0 {
			extConfig, err = resolveExtends(extConfig, newBasePath, visited)
			if err != nil {
				return nil, err
			}
		}

		// Merge extended config into base
		mergeConfigs(merged, extConfig)
	}

	// Finally, merge the local config (highest priority)
	localConfig := *config
	localConfig.Extends = nil // Clear extends to avoid infinite loop
	mergeConfigs(merged, &localConfig)

	return merged, nil
}

// normalizeRef normalizes a reference for comparison
func normalizeRef(ref, basePath string) string {
	if isURL(ref) {
		return ref
	}
	if filepath.IsAbs(ref) {
		return ref
	}
	return filepath.Join(basePath, ref)
}

// isURL checks if a string is a URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// loadExtendedConfig loads a config from URL or file path
func loadExtendedConfig(ref, basePath string) (*Config, string, error) {
	if isURL(ref) {
		config, err := loadFromURL(ref)
		return config, "", err
	}

	// Resolve relative path
	var filePath string
	if filepath.IsAbs(ref) {
		filePath = ref
	} else {
		filePath = filepath.Join(basePath, ref)
	}

	config, err := loadSingleFile(filePath)
	return config, filepath.Dir(filePath), err
}

// loadFromURL loads a config from a URL
func loadFromURL(url string) (*Config, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch %s: status %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from %s: %w", url, err)
	}

	var config Config
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		if err == io.EOF {
			return &config, nil
		}
		return nil, fmt.Errorf("failed to parse config from %s: %w", url, err)
	}

	return &config, nil
}
