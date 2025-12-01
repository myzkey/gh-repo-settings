package config

import (
	"testing"
)

func TestValidateEnvName(t *testing.T) {
	tests := []struct {
		name    string
		envName string
		kind    string
		wantErr bool
	}{
		// Valid names
		{
			name:    "simple uppercase",
			envName: "MY_VAR",
			kind:    "variable",
			wantErr: false,
		},
		{
			name:    "simple lowercase",
			envName: "my_var",
			kind:    "variable",
			wantErr: false,
		},
		{
			name:    "mixed case",
			envName: "My_Var",
			kind:    "variable",
			wantErr: false,
		},
		{
			name:    "starts with underscore",
			envName: "_MY_VAR",
			kind:    "variable",
			wantErr: false,
		},
		{
			name:    "single letter",
			envName: "A",
			kind:    "variable",
			wantErr: false,
		},
		{
			name:    "single underscore",
			envName: "_",
			kind:    "variable",
			wantErr: false,
		},
		{
			name:    "with numbers",
			envName: "VAR_123",
			kind:    "variable",
			wantErr: false,
		},
		// Invalid names
		{
			name:    "empty name",
			envName: "",
			kind:    "variable",
			wantErr: true,
		},
		{
			name:    "starts with number",
			envName: "1VAR",
			kind:    "variable",
			wantErr: true,
		},
		{
			name:    "contains dash",
			envName: "MY-VAR",
			kind:    "variable",
			wantErr: true,
		},
		{
			name:    "contains space",
			envName: "MY VAR",
			kind:    "variable",
			wantErr: true,
		},
		{
			name:    "contains dot",
			envName: "MY.VAR",
			kind:    "variable",
			wantErr: true,
		},
		{
			name:    "GITHUB_ prefix uppercase",
			envName: "GITHUB_TOKEN",
			kind:    "secret",
			wantErr: true,
		},
		{
			name:    "github_ prefix lowercase",
			envName: "github_token",
			kind:    "secret",
			wantErr: true,
		},
		{
			name:    "GITHUB prefix without underscore is ok",
			envName: "GITHUBTOKEN",
			kind:    "variable",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnvName(tt.envName, tt.kind)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateEnvName(%q, %q) expected error, got nil", tt.envName, tt.kind)
				}
			} else {
				if err != nil {
					t.Errorf("validateEnvName(%q, %q) unexpected error: %v", tt.envName, tt.kind, err)
				}
			}
		})
	}
}

func TestEnvConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		env     *EnvConfig
		wantErr bool
	}{
		{
			name: "valid config",
			env: &EnvConfig{
				Variables: map[string]string{
					"API_URL": "https://api.example.com",
					"DEBUG":   "true",
				},
				Secrets: []string{"API_KEY", "DB_PASSWORD"},
			},
			wantErr: false,
		},
		{
			name: "empty config is valid",
			env: &EnvConfig{
				Variables: map[string]string{},
				Secrets:   []string{},
			},
			wantErr: false,
		},
		{
			name:    "nil maps is valid",
			env:     &EnvConfig{},
			wantErr: false,
		},
		{
			name: "invalid variable name",
			env: &EnvConfig{
				Variables: map[string]string{
					"INVALID-NAME": "value",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid secret name",
			env: &EnvConfig{
				Secrets: []string{"INVALID-SECRET"},
			},
			wantErr: true,
		},
		{
			name: "reserved GITHUB_ variable",
			env: &EnvConfig{
				Variables: map[string]string{
					"GITHUB_TOKEN": "value",
				},
			},
			wantErr: true,
		},
		{
			name: "reserved GITHUB_ secret",
			env: &EnvConfig{
				Secrets: []string{"GITHUB_SECRET"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "empty config",
			config:  &Config{},
			wantErr: false,
		},
		{
			name: "config with valid env",
			config: &Config{
				Env: &EnvConfig{
					Variables: map[string]string{"MY_VAR": "value"},
					Secrets:   []string{"MY_SECRET"},
				},
			},
			wantErr: false,
		},
		{
			name: "config with invalid env",
			config: &Config{
				Env: &EnvConfig{
					Variables: map[string]string{"1INVALID": "value"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
