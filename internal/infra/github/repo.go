package github

import (
	"context"
	"fmt"
)

// GetRepo fetches repository settings
func (c *Client) GetRepo(ctx context.Context) (*RepoData, error) {
	var data RepoData
	if err := c.getJSON(ctx, c.repoPath(""), &data); err != nil {
		return nil, fmt.Errorf("failed to get repo: %w", err)
	}
	return &data, nil
}

// UpdateRepo updates repository settings
func (c *Client) UpdateRepo(ctx context.Context, settings map[string]interface{}) error {
	endpoint := c.repoPath("")

	// Try JSON PATCH first
	if _, err := c.callJSON(ctx, httpPatch, endpoint, settings); err == nil {
		return nil
	}

	// Fallback: use field flags instead
	var extraArgs []string
	for k, v := range settings {
		switch val := v.(type) {
		case string:
			extraArgs = append(extraArgs, "-f", fmt.Sprintf("%s=%s", k, val))
		case bool:
			extraArgs = append(extraArgs, "-F", fmt.Sprintf("%s=%t", k, val))
		}
	}

	_, err := c.callAPI(ctx, httpPatch, endpoint, nil, extraArgs...)
	return err
}

// SetTopics sets repository topics
func (c *Client) SetTopics(ctx context.Context, topics []string) error {
	payload := struct {
		Names []string `json:"names"`
	}{Names: topics}
	_, err := c.callJSON(ctx, httpPut, c.repoPath("topics"), payload)
	return err
}
