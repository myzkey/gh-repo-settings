package diff

import (
	"context"
	"testing"

	"github.com/myzkey/gh-repo-settings/internal/config"
	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
	"github.com/myzkey/gh-repo-settings/internal/infra/github"
)

func TestCalculatorErrors(t *testing.T) {
	t.Run("GetRepo error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetRepoError = apperrors.ErrRepoNotFound

		cfg := &config.Config{Repo: &config.RepoConfig{Description: ptr("test")}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.Calculate(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetLabels error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetLabelsError = apperrors.ErrPermissionDenied

		cfg := &config.Config{Labels: &config.LabelsConfig{Items: []config.Label{{Name: "bug", Color: "d73a4a"}}}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.Calculate(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetBranchProtection error (not ErrBranchNotProtected)", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetBranchProtectionError = apperrors.ErrPermissionDenied

		cfg := &config.Config{BranchProtection: map[string]*config.BranchRule{"main": {RequiredReviews: ptr(1)}}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.Calculate(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetSecrets error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetSecretsError = apperrors.ErrPermissionDenied

		cfg := &config.Config{Env: &config.EnvConfig{Secrets: []string{"KEY"}}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{CheckSecrets: true})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetVariables error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetVariablesError = apperrors.ErrPermissionDenied

		cfg := &config.Config{Env: &config.EnvConfig{Variables: map[string]string{"VAR": "value"}}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.CalculateWithOptions(context.Background(), CalculateOptions{CheckEnv: true})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetActionsPermissions error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.GetActionsPermissionsError = apperrors.ErrPermissionDenied

		cfg := &config.Config{Actions: &config.ActionsConfig{Enabled: ptr(true)}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.Calculate(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetActionsWorkflowPermissions error", func(t *testing.T) {
		mock := github.NewMockClient()
		mock.ActionsPermissions = &github.ActionsPermissionsData{Enabled: true}
		mock.GetActionsWorkflowPermissionsError = apperrors.ErrPermissionDenied

		cfg := &config.Config{Actions: &config.ActionsConfig{Enabled: ptr(true)}}
		calc := NewCalculator(mock, cfg)

		_, err := calc.Calculate(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})
}
