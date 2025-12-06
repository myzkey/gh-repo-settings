package github

import (
	"context"
	"fmt"

	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
)

// GetPages fetches GitHub Pages configuration
func (c *Client) GetPages(ctx context.Context) (*PagesData, error) {
	var data PagesData
	err := c.getJSON(ctx, c.repoPath("pages"), &data)
	if err != nil {
		// Pages not enabled returns 404
		var apiErr *apperrors.APIError
		if apperrors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			return nil, apperrors.ErrPagesNotEnabled
		}
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}
	return &data, nil
}

// CreatePages creates GitHub Pages for the repository
func (c *Client) CreatePages(ctx context.Context, buildType string, source *PagesSourceData) error {
	payload := map[string]interface{}{
		"build_type": buildType,
	}
	if source != nil && buildType == "legacy" {
		payload["source"] = map[string]string{
			"branch": source.Branch,
			"path":   source.Path,
		}
	}

	_, err := c.callJSON(ctx, httpPost, c.repoPath("pages"), payload)
	return err
}

// UpdatePages updates GitHub Pages configuration
func (c *Client) UpdatePages(ctx context.Context, buildType string, source *PagesSourceData) error {
	payload := map[string]interface{}{
		"build_type": buildType,
	}
	if source != nil && buildType == "legacy" {
		payload["source"] = map[string]string{
			"branch": source.Branch,
			"path":   source.Path,
		}
	}

	_, err := c.callJSON(ctx, httpPut, c.repoPath("pages"), payload)
	return err
}
