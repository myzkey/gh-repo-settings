package github

import (
	"context"
	"fmt"
)

// GetActionsPermissions fetches Actions permissions for the repository
func (c *Client) GetActionsPermissions(ctx context.Context) (*ActionsPermissionsData, error) {
	var data ActionsPermissionsData
	if err := c.getJSON(ctx, c.repoPath("actions/permissions"), &data); err != nil {
		return nil, fmt.Errorf("failed to get actions permissions: %w", err)
	}
	return &data, nil
}

// UpdateActionsPermissions updates Actions permissions for the repository
func (c *Client) UpdateActionsPermissions(ctx context.Context, enabled bool, allowedActions string) error {
	payload := map[string]interface{}{
		"enabled": enabled,
	}
	if enabled && allowedActions != "" {
		payload["allowed_actions"] = allowedActions
	}

	_, err := c.callJSON(ctx, httpPut, c.repoPath("actions/permissions"), payload)
	return err
}

// GetActionsSelectedActions fetches selected actions configuration
func (c *Client) GetActionsSelectedActions(ctx context.Context) (*ActionsSelectedData, error) {
	var data ActionsSelectedData
	if err := c.getJSON(ctx, c.repoPath("actions/permissions/selected-actions"), &data); err != nil {
		return nil, fmt.Errorf("failed to get selected actions: %w", err)
	}
	return &data, nil
}

// UpdateActionsSelectedActions updates selected actions configuration
func (c *Client) UpdateActionsSelectedActions(ctx context.Context, settings *ActionsSelectedData) error {
	_, err := c.callJSON(ctx, httpPut, c.repoPath("actions/permissions/selected-actions"), settings)
	return err
}

// GetActionsWorkflowPermissions fetches workflow permissions
func (c *Client) GetActionsWorkflowPermissions(ctx context.Context) (*ActionsWorkflowPermissionsData, error) {
	var data ActionsWorkflowPermissionsData
	if err := c.getJSON(ctx, c.repoPath("actions/permissions/workflow"), &data); err != nil {
		return nil, fmt.Errorf("failed to get workflow permissions: %w", err)
	}
	return &data, nil
}

// UpdateActionsWorkflowPermissions updates workflow permissions
func (c *Client) UpdateActionsWorkflowPermissions(ctx context.Context, permissions string, canApprove bool) error {
	payload := map[string]interface{}{
		"default_workflow_permissions":     permissions,
		"can_approve_pull_request_reviews": canApprove,
	}
	_, err := c.callJSON(ctx, httpPut, c.repoPath("actions/permissions/workflow"), payload)
	return err
}
