package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
)

// RepoInfo represents repository owner and name
type RepoInfo struct {
	Owner string
	Name  string
}

// Client wraps gh CLI commands
type Client struct {
	Repo RepoInfo
}

// NewClient creates a new GitHub client
func NewClient(repoArg string) (*Client, error) {
	return NewClientWithContext(context.Background(), repoArg)
}

// NewClientWithContext creates a new GitHub client with context
func NewClientWithContext(ctx context.Context, repoArg string) (*Client, error) {
	var repo RepoInfo
	var err error

	if repoArg != "" {
		repo, err = parseRepoArg(repoArg)
	} else {
		repo, err = getCurrentRepo(ctx)
	}

	if err != nil {
		return nil, err
	}

	return &Client{Repo: repo}, nil
}

func parseRepoArg(arg string) (RepoInfo, error) {
	parts := strings.Split(arg, "/")
	if len(parts) != 2 {
		return RepoInfo{}, apperrors.NewValidationError("repo", fmt.Sprintf("invalid format: %s. Expected: owner/name", arg))
	}
	return RepoInfo{Owner: parts[0], Name: parts[1]}, nil
}

func getCurrentRepo(ctx context.Context) (RepoInfo, error) {
	cmd := exec.CommandContext(ctx, "gh", "repo", "view", "--json", "owner,name")
	out, err := cmd.Output()
	if err != nil {
		return RepoInfo{}, apperrors.NewAPIError("GET", "repo/view", 0, "could not determine repository. Use --repo flag or run from a git repository", err)
	}

	var result struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return RepoInfo{}, apperrors.NewAPIError("GET", "repo/view", 0, "failed to parse repo info", err)
	}

	return RepoInfo{Owner: result.Owner.Login, Name: result.Name}, nil
}

// RepoOwner returns the repository owner
func (c *Client) RepoOwner() string {
	return c.Repo.Owner
}

// RepoName returns the repository name
func (c *Client) RepoName() string {
	return c.Repo.Name
}

// ghAPI executes a gh api command and returns the result
func (c *Client) ghAPI(ctx context.Context, endpoint string, args ...string) ([]byte, error) {
	cmdArgs := []string{"api", endpoint}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, "gh", cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, apperrors.NewAPIError("", endpoint, exitErr.ExitCode(), string(exitErr.Stderr), err)
		}
		return nil, apperrors.NewAPIError("", endpoint, 0, err.Error(), err)
	}
	return out, nil
}

// ghAPIWithInput executes a gh api command with stdin input
func (c *Client) ghAPIWithInput(ctx context.Context, endpoint string, input []byte, args ...string) ([]byte, error) {
	cmdArgs := []string{"api", endpoint}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--input", "-")

	cmd := exec.CommandContext(ctx, "gh", cmdArgs...)
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, apperrors.NewAPIError("", endpoint, exitErr.ExitCode(), string(exitErr.Stderr), err)
		}
		return nil, apperrors.NewAPIError("", endpoint, 0, err.Error(), err)
	}
	return out, nil
}

// GetRepo fetches repository settings
func (c *Client) GetRepo(ctx context.Context) (*RepoData, error) {
	endpoint := fmt.Sprintf("repos/%s/%s", c.Repo.Owner, c.Repo.Name)
	out, err := c.ghAPI(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo: %w", err)
	}

	var data RepoData
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("failed to parse repo data: %w", err)
	}

	return &data, nil
}

// UpdateRepo updates repository settings
func (c *Client) UpdateRepo(ctx context.Context, settings map[string]interface{}) error {
	endpoint := fmt.Sprintf("repos/%s/%s", c.Repo.Owner, c.Repo.Name)
	jsonData, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = c.ghAPIWithInput(ctx, endpoint, jsonData, "-X", "PATCH", "-H", "Accept: application/vnd.github+json")
	if err != nil {
		// Try with field flags instead
		args := []string{"api", endpoint, "-X", "PATCH"}
		for k, v := range settings {
			switch val := v.(type) {
			case string:
				args = append(args, "-f", fmt.Sprintf("%s=%s", k, val))
			case bool:
				args = append(args, "-F", fmt.Sprintf("%s=%t", k, val))
			}
		}
		cmd := exec.CommandContext(ctx, "gh", args...)
		_, err = cmd.Output()
		if err != nil {
			return apperrors.NewAPIError("PATCH", endpoint, 0, "failed to update repo", err)
		}
	}

	return nil
}

// GetLabels fetches repository labels
func (c *Client) GetLabels(ctx context.Context) ([]LabelData, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/labels", c.Repo.Owner, c.Repo.Name)
	out, err := c.ghAPI(ctx, endpoint, "--paginate")
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	var labels []LabelData
	if err := json.Unmarshal(out, &labels); err != nil {
		return nil, fmt.Errorf("failed to parse labels: %w", err)
	}

	return labels, nil
}

// CreateLabel creates a new label
func (c *Client) CreateLabel(ctx context.Context, name, color, description string) error {
	endpoint := fmt.Sprintf("repos/%s/%s/labels", c.Repo.Owner, c.Repo.Name)
	args := []string{
		"-X", "POST",
		"-f", fmt.Sprintf("name=%s", name),
		"-f", fmt.Sprintf("color=%s", color),
	}
	if description != "" {
		args = append(args, "-f", fmt.Sprintf("description=%s", description))
	}

	_, err := c.ghAPI(ctx, endpoint, args...)
	return err
}

// UpdateLabel updates an existing label
func (c *Client) UpdateLabel(ctx context.Context, oldName, newName, color, description string) error {
	endpoint := fmt.Sprintf("repos/%s/%s/labels/%s", c.Repo.Owner, c.Repo.Name, oldName)
	args := []string{
		"-X", "PATCH",
		"-f", fmt.Sprintf("new_name=%s", newName),
		"-f", fmt.Sprintf("color=%s", color),
	}
	if description != "" {
		args = append(args, "-f", fmt.Sprintf("description=%s", description))
	}

	_, err := c.ghAPI(ctx, endpoint, args...)
	return err
}

// DeleteLabel deletes a label
func (c *Client) DeleteLabel(ctx context.Context, name string) error {
	endpoint := fmt.Sprintf("repos/%s/%s/labels/%s", c.Repo.Owner, c.Repo.Name, name)
	_, err := c.ghAPI(ctx, endpoint, "-X", "DELETE")
	return err
}

// SetTopics sets repository topics
func (c *Client) SetTopics(ctx context.Context, topics []string) error {
	endpoint := fmt.Sprintf("repos/%s/%s/topics", c.Repo.Owner, c.Repo.Name)

	body := struct {
		Names []string `json:"names"`
	}{Names: topics}
	bodyJSON, _ := json.Marshal(body)

	args := []string{
		"-X", "PUT",
		"-H", "Accept: application/vnd.github+json",
	}

	_, err := c.ghAPIWithInput(ctx, endpoint, bodyJSON, args...)
	return err
}

// GetSecrets fetches repository secret names
func (c *Client) GetSecrets(ctx context.Context) ([]string, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/actions/secrets", c.Repo.Owner, c.Repo.Name)
	out, err := c.ghAPI(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var result struct {
		Secrets []struct {
			Name string `json:"name"`
		} `json:"secrets"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}

	names := make([]string, len(result.Secrets))
	for i, s := range result.Secrets {
		names[i] = s.Name
	}

	return names, nil
}

// GetVariables fetches repository variable names
func (c *Client) GetVariables(ctx context.Context) ([]string, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/actions/variables", c.Repo.Owner, c.Repo.Name)
	out, err := c.ghAPI(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var result struct {
		Variables []struct {
			Name string `json:"name"`
		} `json:"variables"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}

	names := make([]string, len(result.Variables))
	for i, v := range result.Variables {
		names[i] = v.Name
	}

	return names, nil
}

// GetBranchProtection fetches branch protection rules
func (c *Client) GetBranchProtection(ctx context.Context, branch string) (*BranchProtectionData, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/branches/%s/protection", c.Repo.Owner, c.Repo.Name, branch)
	out, err := c.ghAPI(ctx, endpoint)
	if err != nil {
		// Check if branch protection doesn't exist
		if apiErr, ok := err.(*apperrors.APIError); ok && apiErr.StatusCode == 1 {
			return nil, apperrors.ErrBranchNotProtected
		}
		return nil, err
	}

	var data BranchProtectionData
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// UpdateBranchProtection updates branch protection rules
func (c *Client) UpdateBranchProtection(ctx context.Context, branch string, settings *BranchProtectionSettings) error {
	endpoint := fmt.Sprintf("repos/%s/%s/branches/%s/protection", c.Repo.Owner, c.Repo.Name, branch)

	// Build the protection payload
	payload := map[string]interface{}{
		"enforce_admins":          settings.EnforceAdmins != nil && *settings.EnforceAdmins,
		"required_linear_history": settings.RequireLinearHistory != nil && *settings.RequireLinearHistory,
		"allow_force_pushes":      settings.AllowForcePushes != nil && *settings.AllowForcePushes,
		"allow_deletions":         settings.AllowDeletions != nil && *settings.AllowDeletions,
		"restrictions":            nil,
	}

	// Required pull request reviews
	if settings.RequiredReviews != nil || settings.DismissStaleReviews != nil || settings.RequireCodeOwnerReviews != nil {
		reviews := map[string]interface{}{}
		if settings.RequiredReviews != nil {
			reviews["required_approving_review_count"] = *settings.RequiredReviews
		}
		if settings.DismissStaleReviews != nil {
			reviews["dismiss_stale_reviews"] = *settings.DismissStaleReviews
		}
		if settings.RequireCodeOwnerReviews != nil {
			reviews["require_code_owner_reviews"] = *settings.RequireCodeOwnerReviews
		}
		payload["required_pull_request_reviews"] = reviews
	} else {
		payload["required_pull_request_reviews"] = nil
	}

	// Required status checks
	if settings.RequireStatusChecks != nil && *settings.RequireStatusChecks {
		checks := map[string]interface{}{
			"strict": settings.StrictStatusChecks != nil && *settings.StrictStatusChecks,
		}
		if len(settings.StatusChecks) > 0 {
			checks["contexts"] = settings.StatusChecks
		} else {
			checks["contexts"] = []string{}
		}
		payload["required_status_checks"] = checks
	} else {
		payload["required_status_checks"] = nil
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = c.ghAPIWithInput(ctx, endpoint, jsonData, "-X", "PUT", "-H", "Accept: application/vnd.github+json")
	return err
}

// GetActionsPermissions fetches Actions permissions for the repository
func (c *Client) GetActionsPermissions(ctx context.Context) (*ActionsPermissionsData, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/actions/permissions", c.Repo.Owner, c.Repo.Name)
	out, err := c.ghAPI(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions permissions: %w", err)
	}

	var data ActionsPermissionsData
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("failed to parse actions permissions: %w", err)
	}

	return &data, nil
}

// UpdateActionsPermissions updates Actions permissions for the repository
func (c *Client) UpdateActionsPermissions(ctx context.Context, enabled bool, allowedActions string) error {
	endpoint := fmt.Sprintf("repos/%s/%s/actions/permissions", c.Repo.Owner, c.Repo.Name)

	payload := map[string]interface{}{
		"enabled": enabled,
	}
	if enabled && allowedActions != "" {
		payload["allowed_actions"] = allowedActions
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = c.ghAPIWithInput(ctx, endpoint, jsonData, "-X", "PUT", "-H", "Accept: application/vnd.github+json")
	return err
}

// GetActionsSelectedActions fetches selected actions configuration
func (c *Client) GetActionsSelectedActions(ctx context.Context) (*ActionsSelectedData, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/actions/permissions/selected-actions", c.Repo.Owner, c.Repo.Name)
	out, err := c.ghAPI(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get selected actions: %w", err)
	}

	var data ActionsSelectedData
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("failed to parse selected actions: %w", err)
	}

	return &data, nil
}

// UpdateActionsSelectedActions updates selected actions configuration
func (c *Client) UpdateActionsSelectedActions(ctx context.Context, settings *ActionsSelectedData) error {
	endpoint := fmt.Sprintf("repos/%s/%s/actions/permissions/selected-actions", c.Repo.Owner, c.Repo.Name)

	jsonData, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = c.ghAPIWithInput(ctx, endpoint, jsonData, "-X", "PUT", "-H", "Accept: application/vnd.github+json")
	return err
}

// GetActionsWorkflowPermissions fetches workflow permissions
func (c *Client) GetActionsWorkflowPermissions(ctx context.Context) (*ActionsWorkflowPermissionsData, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/actions/permissions/workflow", c.Repo.Owner, c.Repo.Name)
	out, err := c.ghAPI(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow permissions: %w", err)
	}

	var data ActionsWorkflowPermissionsData
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("failed to parse workflow permissions: %w", err)
	}

	return &data, nil
}

// UpdateActionsWorkflowPermissions updates workflow permissions
func (c *Client) UpdateActionsWorkflowPermissions(ctx context.Context, permissions string, canApprove bool) error {
	endpoint := fmt.Sprintf("repos/%s/%s/actions/permissions/workflow", c.Repo.Owner, c.Repo.Name)

	payload := map[string]interface{}{
		"default_workflow_permissions":     permissions,
		"can_approve_pull_request_reviews": canApprove,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = c.ghAPIWithInput(ctx, endpoint, jsonData, "-X", "PUT", "-H", "Accept: application/vnd.github+json")
	return err
}
