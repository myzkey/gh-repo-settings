package comparator

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	"github.com/myzkey/gh-repo-settings/internal/diff/domain/model"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

func TestEnvComparator_CompareSecrets(t *testing.T) {
	tests := []struct {
		name          string
		currentSecrets []string
		configSecrets []string
		dotEnv        *config.DotEnvValues
		syncDelete    bool
		expectAdds    int
		expectMissing int
		expectDeletes int
	}{
		{
			name:           "all secrets present - no changes",
			currentSecrets: []string{"API_KEY", "SECRET"},
			configSecrets:  []string{"API_KEY", "SECRET"},
			expectAdds:     0,
			expectMissing:  0,
			expectDeletes:  0,
		},
		{
			name:           "missing secret without dotenv",
			currentSecrets: []string{},
			configSecrets:  []string{"API_KEY"},
			expectAdds:     0,
			expectMissing:  1,
			expectDeletes:  0,
		},
		{
			name:           "missing secret with dotenv value",
			currentSecrets: []string{},
			configSecrets:  []string{"API_KEY"},
			dotEnv: &config.DotEnvValues{
				Values: map[string]string{"API_KEY": "secret_value"},
			},
			expectAdds:    1,
			expectMissing: 0,
			expectDeletes: 0,
		},
		{
			name:           "delete secret with syncDelete",
			currentSecrets: []string{"API_KEY", "OLD_SECRET"},
			configSecrets:  []string{"API_KEY"},
			syncDelete:     true,
			expectAdds:     0,
			expectMissing:  0,
			expectDeletes:  1,
		},
		{
			name:           "no delete without syncDelete",
			currentSecrets: []string{"API_KEY", "OLD_SECRET"},
			configSecrets:  []string{"API_KEY"},
			syncDelete:     false,
			expectAdds:     0,
			expectMissing:  0,
			expectDeletes:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Secrets = tt.currentSecrets

			comparator := NewEnvComparator(mock, &config.EnvConfig{
				Secrets: tt.configSecrets,
			}, tt.dotEnv, EnvComparatorOptions{
				CheckSecrets: true,
				SyncDelete:   tt.syncDelete,
			})

			plan, err := comparator.Compare(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var adds, missing, deletes int
			for _, c := range plan.Changes() {
				if c.Category != model.CategorySecrets {
					t.Errorf("expected category %s, got %s", model.CategorySecrets, c.Category)
				}
				switch c.Type {
				case model.ChangeAdd:
					adds++
				case model.ChangeMissing:
					missing++
				case model.ChangeDelete:
					deletes++
				}
			}

			if adds != tt.expectAdds {
				t.Errorf("expected %d adds, got %d", tt.expectAdds, adds)
			}
			if missing != tt.expectMissing {
				t.Errorf("expected %d missing, got %d", tt.expectMissing, missing)
			}
			if deletes != tt.expectDeletes {
				t.Errorf("expected %d deletes, got %d", tt.expectDeletes, deletes)
			}
		})
	}
}

func TestEnvComparator_CompareVariables(t *testing.T) {
	tests := []struct {
		name          string
		currentVars   []github.VariableData
		configVars    map[string]string
		dotEnv        *config.DotEnvValues
		syncDelete    bool
		expectAdds    int
		expectUpdates int
		expectDeletes int
	}{
		{
			name:        "no changes when values match",
			currentVars: []github.VariableData{{Name: "ENV", Value: "prod"}},
			configVars:  map[string]string{"ENV": "prod"},
			expectAdds:  0,
			expectUpdates: 0,
			expectDeletes: 0,
		},
		{
			name:        "add new variable",
			currentVars: []github.VariableData{},
			configVars:  map[string]string{"ENV": "prod"},
			expectAdds:  1,
			expectUpdates: 0,
			expectDeletes: 0,
		},
		{
			name:        "update existing variable",
			currentVars: []github.VariableData{{Name: "ENV", Value: "dev"}},
			configVars:  map[string]string{"ENV": "prod"},
			expectAdds:  0,
			expectUpdates: 1,
			expectDeletes: 0,
		},
		{
			name:        "dotenv overrides yaml default",
			currentVars: []github.VariableData{},
			configVars:  map[string]string{"ENV": "yaml_default"},
			dotEnv: &config.DotEnvValues{
				Values: map[string]string{"ENV": "dotenv_value"},
			},
			expectAdds:    1,
			expectUpdates: 0,
			expectDeletes: 0,
		},
		{
			name:          "delete variable with syncDelete",
			currentVars:   []github.VariableData{{Name: "ENV", Value: "prod"}, {Name: "OLD", Value: "x"}},
			configVars:    map[string]string{"ENV": "prod"},
			syncDelete:    true,
			expectAdds:    0,
			expectUpdates: 0,
			expectDeletes: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := github.NewMockClient()
			mock.Variables = tt.currentVars

			comparator := NewEnvComparator(mock, &config.EnvConfig{
				Variables: tt.configVars,
			}, tt.dotEnv, EnvComparatorOptions{
				CheckVars:  true,
				SyncDelete: tt.syncDelete,
			})

			plan, err := comparator.Compare(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var adds, updates, deletes int
			for _, c := range plan.Changes() {
				if c.Category != model.CategoryVariables {
					t.Errorf("expected category %s, got %s", model.CategoryVariables, c.Category)
				}
				switch c.Type {
				case model.ChangeAdd:
					adds++
				case model.ChangeUpdate:
					updates++
				case model.ChangeDelete:
					deletes++
				}
			}

			if adds != tt.expectAdds {
				t.Errorf("expected %d adds, got %d", tt.expectAdds, adds)
			}
			if updates != tt.expectUpdates {
				t.Errorf("expected %d updates, got %d", tt.expectUpdates, updates)
			}
			if deletes != tt.expectDeletes {
				t.Errorf("expected %d deletes, got %d", tt.expectDeletes, deletes)
			}
		})
	}
}

func TestEnvComparator_OptionsFlags(t *testing.T) {
	t.Run("CheckSecrets=false skips secrets comparison", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.Secrets = []string{}

		comparator := NewEnvComparator(mock, &config.EnvConfig{
			Secrets: []string{"API_KEY"},
		}, nil, EnvComparatorOptions{
			CheckSecrets: false,
		})

		plan, err := comparator.Compare(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if plan.HasChanges() {
			t.Error("expected no changes when CheckSecrets=false")
		}
	})

	t.Run("CheckVars=false skips variables comparison", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.Variables = []github.VariableData{}

		comparator := NewEnvComparator(mock, &config.EnvConfig{
			Variables: map[string]string{"ENV": "prod"},
		}, nil, EnvComparatorOptions{
			CheckVars: false,
		})

		plan, err := comparator.Compare(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if plan.HasChanges() {
			t.Error("expected no changes when CheckVars=false")
		}
	})
}

func TestEnvComparator_Errors(t *testing.T) {
	t.Run("GetSecrets error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetSecretsError = apperrors.ErrPermissionDenied

		comparator := NewEnvComparator(mock, &config.EnvConfig{
			Secrets: []string{"KEY"},
		}, nil, EnvComparatorOptions{CheckSecrets: true})

		_, err := comparator.Compare(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("GetVariables error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetVariablesError = apperrors.ErrPermissionDenied

		comparator := NewEnvComparator(mock, &config.EnvConfig{
			Variables: map[string]string{"VAR": "val"},
		}, nil, EnvComparatorOptions{CheckVars: true})

		_, err := comparator.Compare(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
