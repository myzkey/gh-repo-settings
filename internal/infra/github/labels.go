package github

import (
	"context"
	"fmt"
)

// GetLabels fetches repository labels
func (c *Client) GetLabels(ctx context.Context) ([]LabelData, error) {
	var labels []LabelData
	if err := c.getJSON(ctx, c.repoPath("labels"), &labels, "--paginate"); err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}
	return labels, nil
}

// CreateLabel creates a new label
func (c *Client) CreateLabel(ctx context.Context, name, color, description string) error {
	payload := map[string]string{
		"name":  name,
		"color": color,
	}
	if description != "" {
		payload["description"] = description
	}
	_, err := c.callJSON(ctx, httpPost, c.repoPath("labels"), payload)
	return err
}

// UpdateLabel updates an existing label
func (c *Client) UpdateLabel(ctx context.Context, oldName, newName, color, description string) error {
	payload := map[string]string{
		"new_name": newName,
		"color":    color,
	}
	if description != "" {
		payload["description"] = description
	}
	_, err := c.callJSON(ctx, httpPatch, c.repoPath(labelPath(oldName)), payload)
	return err
}

// DeleteLabel deletes a label
func (c *Client) DeleteLabel(ctx context.Context, name string) error {
	_, err := c.callAPI(ctx, httpDelete, c.repoPath(labelPath(name)), nil)
	return err
}
