package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	apperrors "github.com/myzkey/gh-repo-settings/internal/errors"
)

// httpStatusRegex matches "HTTP XXX" in gh api stderr output
var httpStatusRegex = regexp.MustCompile(`HTTP (\d{3})`)

// httpMethod represents an HTTP method for API calls
type httpMethod string

const (
	httpGet    httpMethod = "GET"
	httpPost   httpMethod = "POST"
	httpPut    httpMethod = "PUT"
	httpPatch  httpMethod = "PATCH"
	httpDelete httpMethod = "DELETE"
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

// repoPath builds an API endpoint path for the current repository.
// Example: repoPath("labels") returns "repos/{owner}/{name}/labels"
func (c *Client) repoPath(path string) string {
	if path == "" {
		return fmt.Sprintf("repos/%s/%s", c.Repo.Owner, c.Repo.Name)
	}
	return fmt.Sprintf("repos/%s/%s/%s", c.Repo.Owner, c.Repo.Name, path)
}

// branchPath builds an API endpoint path for branch-related operations.
// It URL-encodes the branch name to handle branches with slashes (e.g., "feature/foo").
// Example: branchPath("main", "protection") returns "repos/{owner}/{name}/branches/main/protection"
func (c *Client) branchPath(branch, suffix string) string {
	encodedBranch := url.PathEscape(branch)
	if suffix == "" {
		return c.repoPath(fmt.Sprintf("branches/%s", encodedBranch))
	}
	return c.repoPath(fmt.Sprintf("branches/%s/%s", encodedBranch, suffix))
}

// labelPath builds an API endpoint path for label operations.
// It URL-encodes the label name to handle labels with spaces, Japanese characters, or special symbols.
// Example: labelPath("bug") returns "labels/bug", labelPath("help wanted") returns "labels/help%20wanted"
func labelPath(name string) string {
	return "labels/" + url.PathEscape(name)
}

// secretPath builds an API endpoint path for secret operations.
// It URL-encodes the secret name to handle names with special characters.
// Example: secretPath("MY_SECRET") returns "actions/secrets/MY_SECRET"
func secretPath(name string) string {
	return "actions/secrets/" + url.PathEscape(name)
}

// variablePath builds an API endpoint path for variable operations.
// It URL-encodes the variable name to handle names with special characters.
// Example: variablePath("MY_VAR") returns "actions/variables/MY_VAR"
func variablePath(name string) string {
	return "actions/variables/" + url.PathEscape(name)
}

// parseHTTPStatus extracts HTTP status code from gh api stderr output
// Returns 0 if no status code is found
func parseHTTPStatus(stderr string) int {
	matches := httpStatusRegex.FindStringSubmatch(stderr)
	if len(matches) >= 2 {
		if code, err := strconv.Atoi(matches[1]); err == nil {
			return code
		}
	}
	return 0
}

// callAPI is the low-level function for executing gh api commands.
// It handles GET requests (body must be nil) and other methods with optional body data.
func (c *Client) callAPI(ctx context.Context, method httpMethod, endpoint string, body []byte, extraArgs ...string) ([]byte, error) {
	if method == httpGet && body != nil {
		return nil, fmt.Errorf("GET request must not have body")
	}

	cmdArgs := []string{"api", endpoint}
	if method != httpGet {
		cmdArgs = append(cmdArgs, "-X", string(method))
	}
	cmdArgs = append(cmdArgs, extraArgs...)

	var cmd *exec.Cmd
	if body != nil {
		cmdArgs = append(cmdArgs, "--input", "-")
		cmd = exec.CommandContext(ctx, "gh", cmdArgs...)
		cmd.Stdin = bytes.NewReader(body)
	} else {
		cmd = exec.CommandContext(ctx, "gh", cmdArgs...)
	}

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			statusCode := parseHTTPStatus(stderr)
			return nil, apperrors.NewAPIError(string(method), endpoint, statusCode, stderr, err)
		}
		return nil, apperrors.NewAPIError(string(method), endpoint, 0, err.Error(), err)
	}
	return out, nil
}

// jsonHeaders returns the standard headers for JSON API requests
func jsonHeaders() []string {
	return []string{"-H", "Accept: application/vnd.github+json"}
}

// getJSON performs a GET request to the given endpoint and unmarshals the JSON response into result.
// This function is GET-only; use callJSON for POST/PUT/PATCH/DELETE requests.
func (c *Client) getJSON(ctx context.Context, endpoint string, result interface{}, extraArgs ...string) error {
	out, err := c.callAPI(ctx, httpGet, endpoint, nil, extraArgs...)
	if err != nil {
		return err
	}
	return json.Unmarshal(out, result)
}

// callJSON sends a JSON request body to an endpoint.
// It marshals the body, adds JSON headers, and returns the response.
func (c *Client) callJSON(ctx context.Context, method httpMethod, endpoint string, body interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body for %s: %w", endpoint, err)
	}
	return c.callAPI(ctx, method, endpoint, jsonData, jsonHeaders()...)
}
