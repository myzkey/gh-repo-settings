package diff

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

func TestCalculatorCheckSecrets(t *testing.T) {
	tests := []struct {
		name           string
		currentSecrets []string
		configSecrets  []string
		expectMissing  int
	}{
		{
			name:           "all secrets present",
			currentSecrets: []string{"API_KEY", "DEPLOY_TOKEN"},
			configSecrets:  []string{"API_KEY", "DEPLOY_TOKEN"},
			expectMissing:  0,
		},
		{
			name:           "some secrets missing",
			currentSecrets: []string{"API_KEY"},
			configSecrets:  []string{"API_KEY", "DEPLOY_TOKEN", "SECRET_KEY"},
			expectMissing:  2,
		},
		{
			name:           "all secrets missing",
			currentSecrets: []string{},
			configSecrets:  []string{"API_KEY"},
			expectMissing:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Secrets = tt.currentSecrets

			cfg := &config.Config{
				Env: &config.EnvConfig{
					Secrets: tt.configSecrets,
				},
			}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{
				CheckSecrets: true,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			missingCount := 0
			for _, c := range plan.Changes() {
				if c.Category == "secrets" && c.Type == ChangeMissing {
					missingCount++
				}
			}

			if missingCount != tt.expectMissing {
				t.Errorf("expected %d missing secrets, got %d", tt.expectMissing, missingCount)
			}
		})
	}
}

func TestCalculatorCheckVariables(t *testing.T) {
	tests := []struct {
		name        string
		currentVars []github.VariableData
		configVars  map[string]string
		expectAdds  int
	}{
		{
			name:        "all variables present with same values",
			currentVars: []github.VariableData{{Name: "NODE_ENV", Value: "prod"}, {Name: "DEBUG", Value: "true"}},
			configVars:  map[string]string{"NODE_ENV": "prod"},
			expectAdds:  0,
		},
		{
			name:        "some variables to add",
			currentVars: []github.VariableData{{Name: "NODE_ENV", Value: "prod"}},
			configVars:  map[string]string{"NODE_ENV": "prod", "LOG_LEVEL": "info"},
			expectAdds:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Variables = tt.currentVars

			cfg := &config.Config{
				Env: &config.EnvConfig{
					Variables: tt.configVars,
				},
			}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{
				CheckEnv: true,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			addCount := 0
			for _, c := range plan.Changes() {
				if c.Category == "variables" && c.Type == ChangeAdd {
					addCount++
				}
			}

			if addCount != tt.expectAdds {
				t.Errorf("expected %d adds, got %d", tt.expectAdds, addCount)
			}
		})
	}
}

func TestCalculatorCompareEnvSyncDelete(t *testing.T) {
	tests := []struct {
		name           string
		currentSecrets []string
		currentVars    []github.VariableData
		configSecrets  []string
		configVars     map[string]string
		syncDelete     bool
		expectDeletes  int
	}{
		{
			name:           "delete secrets with syncDelete",
			currentSecrets: []string{"KEEP", "DELETE_ME"},
			configSecrets:  []string{"KEEP"},
			syncDelete:     true,
			expectDeletes:  1,
		},
		{
			name:           "no delete without syncDelete",
			currentSecrets: []string{"KEEP", "DELETE_ME"},
			configSecrets:  []string{"KEEP"},
			syncDelete:     false,
			expectDeletes:  0,
		},
		{
			name:          "delete variables with syncDelete",
			currentVars:   []github.VariableData{{Name: "KEEP", Value: "v1"}, {Name: "DELETE_ME", Value: "v2"}},
			configVars:    map[string]string{"KEEP": "v1"},
			syncDelete:    true,
			expectDeletes: 1,
		},
		{
			name:          "no delete variables without syncDelete",
			currentVars:   []github.VariableData{{Name: "KEEP", Value: "v1"}, {Name: "DELETE_ME", Value: "v2"}},
			configVars:    map[string]string{"KEEP": "v1"},
			syncDelete:    false,
			expectDeletes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Secrets = tt.currentSecrets
			mock.Variables = tt.currentVars

			cfg := &config.Config{
				Env: &config.EnvConfig{
					Secrets:   tt.configSecrets,
					Variables: tt.configVars,
				},
			}
			calc := NewCalculator(mock, cfg)

			plan, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{
				CheckSecrets: len(tt.configSecrets) > 0,
				CheckEnv:     len(tt.configVars) > 0,
				SyncDelete:   tt.syncDelete,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			deleteCount := 0
			for _, c := range plan.Changes() {
				if c.Type == ChangeDelete {
					deleteCount++
				}
			}

			if deleteCount != tt.expectDeletes {
				t.Errorf("expected %d deletes, got %d", tt.expectDeletes, deleteCount)
			}
		})
	}
}

func TestCalculatorCompareEnvWithDotEnv(t *testing.T) {
	mock := github.NewMockClient()
	mock.Secrets = []string{}
	mock.Variables = []github.VariableData{}

	dotEnv := &config.DotEnvValues{
		Values: map[string]string{
			"SECRET1": "secret_value",
			"VAR1":    "env_value",
		},
	}

	cfg := &config.Config{
		Env: &config.EnvConfig{
			Secrets:   []string{"SECRET1", "SECRET2"},
			Variables: map[string]string{"VAR1": "yaml_default", "VAR2": "yaml_only"},
		},
	}
	calc := NewCalculatorWithEnv(mock, cfg, dotEnv)

	plan, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{
		CheckSecrets: true,
		CheckEnv:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check SECRET1 (in .env) - should be add, not missing
	// Check SECRET2 (not in .env) - should be missing
	// Check VAR1 - should use env_value from .env
	// Check VAR2 - should use yaml_only

	var secret1Found, secret2Found, var1Found, var2Found bool
	for _, c := range plan.Changes() {
		switch c.Key {
		case "SECRET1":
			secret1Found = true
			if c.Type != ChangeAdd {
				t.Errorf("SECRET1 should be add, got %v", c.Type)
			}
		case "SECRET2":
			secret2Found = true
			if c.Type != ChangeMissing {
				t.Errorf("SECRET2 should be missing, got %v", c.Type)
			}
		case "VAR1":
			var1Found = true
			if c.New != "env_value" {
				t.Errorf("VAR1 should use env_value, got %v", c.New)
			}
		case "VAR2":
			var2Found = true
			if c.New != "yaml_only" {
				t.Errorf("VAR2 should use yaml_only, got %v", c.New)
			}
		}
	}

	if !secret1Found {
		t.Error("SECRET1 change not found")
	}
	if !secret2Found {
		t.Error("SECRET2 change not found")
	}
	if !var1Found {
		t.Error("VAR1 change not found")
	}
	if !var2Found {
		t.Error("VAR2 change not found")
	}
}

func TestCalculatorCompareVariablesUpdate(t *testing.T) {
	mock := github.NewMockClient()
	mock.Variables = []github.VariableData{
		{Name: "NODE_ENV", Value: "development"},
	}

	cfg := &config.Config{
		Env: &config.EnvConfig{
			Variables: map[string]string{"NODE_ENV": "production"},
		},
	}
	calc := NewCalculator(mock, cfg)

	plan, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{
		CheckEnv: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, c := range plan.Changes() {
		if c.Key == "NODE_ENV" && c.Type == ChangeUpdate {
			found = true
			if c.Old != "development" {
				t.Errorf("expected old value 'development', got %v", c.Old)
			}
			if c.New != "production" {
				t.Errorf("expected new value 'production', got %v", c.New)
			}
		}
	}

	if !found {
		t.Error("NODE_ENV update change not found")
	}
}
