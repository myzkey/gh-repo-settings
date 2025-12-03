package provider

import (
	"encoding/json"
	"testing"
)

func TestNewSecretsManagerProvider(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *Config
		wantErr    bool
		errMsg     string
		wantSecret string
		wantRegion string
	}{
		{
			name: "valid config with secret and region",
			cfg: &Config{
				Name:   "secretsmanager",
				Secret: "/myapp/prod/secrets",
				Region: "ap-northeast-1",
			},
			wantErr:    false,
			wantSecret: "/myapp/prod/secrets",
			wantRegion: "ap-northeast-1",
		},
		{
			name: "valid config without region",
			cfg: &Config{
				Name:   "secretsmanager",
				Secret: "/myapp/secrets",
			},
			wantErr:    false,
			wantSecret: "/myapp/secrets",
			wantRegion: "",
		},
		{
			name: "missing secret",
			cfg: &Config{
				Name:   "secretsmanager",
				Region: "us-east-1",
			},
			wantErr: true,
			errMsg:  "secret is required for secretsmanager provider",
		},
		{
			name: "empty secret",
			cfg: &Config{
				Name:   "secretsmanager",
				Secret: "",
				Region: "us-east-1",
			},
			wantErr: true,
			errMsg:  "secret is required for secretsmanager provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewSecretsManagerProvider(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewSecretsManagerProvider() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewSecretsManagerProvider() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("NewSecretsManagerProvider() unexpected error = %v", err)
				return
			}
			if provider.secret != tt.wantSecret {
				t.Errorf("NewSecretsManagerProvider() secret = %v, want %v", provider.secret, tt.wantSecret)
			}
			if provider.region != tt.wantRegion {
				t.Errorf("NewSecretsManagerProvider() region = %v, want %v", provider.region, tt.wantRegion)
			}
		})
	}
}

func TestSecretsManagerProvider_Name(t *testing.T) {
	provider := &SecretsManagerProvider{}
	if got := provider.Name(); got != "secretsmanager" {
		t.Errorf("SecretsManagerProvider.Name() = %v, want %v", got, "secretsmanager")
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:  "string value",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "float64 value",
			input: float64(123.45),
			want:  "123.45",
		},
		{
			name:  "integer as float64",
			input: float64(42),
			want:  "42",
		},
		{
			name:  "bool true",
			input: true,
			want:  "true",
		},
		{
			name:  "bool false",
			input: false,
			want:  "false",
		},
		{
			name:  "map value",
			input: map[string]interface{}{"key": "value"},
			want:  `{"key":"value"}`,
		},
		{
			name:  "array value",
			input: []interface{}{"a", "b"},
			want:  `["a","b"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("toString() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("toString() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("toString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractFromJSON(t *testing.T) {
	jsonStr := `{"API_TOKEN": "secret123", "DB_PASSWORD": "dbpass", "PORT": 8080, "ENABLED": true}`

	tests := []struct {
		name    string
		json    string
		keys    []string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "extract all keys (empty keys)",
			json: jsonStr,
			keys: []string{},
			want: map[string]string{
				"API_TOKEN":   "secret123",
				"DB_PASSWORD": "dbpass",
				"PORT":        "8080",
				"ENABLED":     "true",
			},
		},
		{
			name: "extract specific keys",
			json: jsonStr,
			keys: []string{"API_TOKEN", "DB_PASSWORD"},
			want: map[string]string{
				"API_TOKEN":   "secret123",
				"DB_PASSWORD": "dbpass",
			},
		},
		{
			name: "extract single key",
			json: jsonStr,
			keys: []string{"API_TOKEN"},
			want: map[string]string{
				"API_TOKEN": "secret123",
			},
		},
		{
			name:    "invalid json",
			json:    "not valid json",
			keys:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]interface{}
			err := json.Unmarshal([]byte(tt.json), &data)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for invalid JSON")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			result := make(map[string]string)
			if len(tt.keys) == 0 {
				// Extract all keys
				for k, v := range data {
					strVal, _ := toString(v)
					result[k] = strVal
				}
			} else {
				// Extract specific keys
				for _, key := range tt.keys {
					if v, ok := data[key]; ok {
						strVal, _ := toString(v)
						result[key] = strVal
					}
				}
			}

			if len(result) != len(tt.want) {
				t.Errorf("got %d keys, want %d", len(result), len(tt.want))
			}
			for k, v := range tt.want {
				if result[k] != v {
					t.Errorf("key %q = %q, want %q", k, result[k], v)
				}
			}
		})
	}
}
