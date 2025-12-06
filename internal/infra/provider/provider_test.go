package provider

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
			errMsg:  "provider config is nil",
		},
		{
			name: "secretsmanager provider",
			cfg: &Config{
				Name:   "secretsmanager",
				Secret: "/myapp/secrets",
				Region: "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "secretsmanager without secret",
			cfg: &Config{
				Name:   "secretsmanager",
				Region: "us-east-1",
			},
			wantErr: true,
			errMsg:  "secret is required for secretsmanager provider",
		},
		{
			name: "unknown provider",
			cfg: &Config{
				Name: "unknown",
			},
			wantErr: true,
			errMsg:  "unknown provider: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Errorf("New() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("New() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("New() unexpected error = %v", err)
				return
			}
			if got == nil {
				t.Error("New() returned nil provider")
			}
		})
	}
}

func TestLoadSecrets(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		keys    []string
		wantErr bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			keys:    []string{"key1"},
			wantErr: true,
		},
		{
			name: "unknown provider",
			cfg: &Config{
				Name: "unknown",
			},
			keys:    []string{"key1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadSecrets(context.Background(), tt.cfg, tt.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSecrets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
