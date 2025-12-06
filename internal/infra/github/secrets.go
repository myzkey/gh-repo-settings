package github

import (
	"context"
	"fmt"
	"os/exec"

	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
)

// GetSecrets fetches repository secret names
func (c *Client) GetSecrets(ctx context.Context) ([]string, error) {
	var result struct {
		Secrets []struct {
			Name string `json:"name"`
		} `json:"secrets"`
	}
	if err := c.getJSON(ctx, c.repoPath("actions/secrets"), &result, "--paginate"); err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}

	names := make([]string, len(result.Secrets))
	for i, s := range result.Secrets {
		names[i] = s.Name
	}
	return names, nil
}

// SetSecret creates or updates a repository secret using gh secret set
func (c *Client) SetSecret(ctx context.Context, name, value string) error {
	repo := fmt.Sprintf("%s/%s", c.Repo.Owner, c.Repo.Name)
	cmd := exec.CommandContext(ctx, "gh", "secret", "set", name, "--repo", repo, "--body", value)
	_, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return apperrors.NewAPIError("SET", "secret/"+name, exitErr.ExitCode(), string(exitErr.Stderr), err)
		}
		return err
	}
	return nil
}

// DeleteSecret deletes a repository secret
func (c *Client) DeleteSecret(ctx context.Context, name string) error {
	_, err := c.callAPI(ctx, httpDelete, c.repoPath(secretPath(name)), nil)
	return err
}

// GetVariables fetches repository variables with their values
func (c *Client) GetVariables(ctx context.Context) ([]VariableData, error) {
	var result struct {
		Variables []VariableData `json:"variables"`
	}
	if err := c.getJSON(ctx, c.repoPath("actions/variables"), &result, "--paginate"); err != nil {
		return nil, fmt.Errorf("failed to get variables: %w", err)
	}
	return result.Variables, nil
}

// SetVariable creates or updates a repository variable
func (c *Client) SetVariable(ctx context.Context, name, value string) error {
	// First, try to get the variable to see if it exists
	varEndpoint := c.repoPath(variablePath(name))
	_, getErr := c.callAPI(ctx, httpGet, varEndpoint, nil)

	payload := map[string]string{
		"name":  name,
		"value": value,
	}

	if getErr != nil {
		// Check if it's a 404 (not found) error
		var apiErr *apperrors.APIError
		if apperrors.As(getErr, &apiErr) && apiErr.StatusCode == 404 {
			// Variable doesn't exist, create it
			_, err := c.callJSON(ctx, httpPost, c.repoPath("actions/variables"), payload)
			return err
		}
		// Other error (permission denied, rate limited, etc.)
		return fmt.Errorf("failed to check variable existence: %w", getErr)
	}

	// Variable exists, update it
	_, err := c.callJSON(ctx, httpPatch, varEndpoint, payload)
	return err
}

// DeleteVariable deletes a repository variable
func (c *Client) DeleteVariable(ctx context.Context, name string) error {
	_, err := c.callAPI(ctx, httpDelete, c.repoPath(variablePath(name)), nil)
	return err
}
