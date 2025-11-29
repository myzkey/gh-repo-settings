package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
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
	var repo RepoInfo
	var err error

	if repoArg != "" {
		repo, err = parseRepoArg(repoArg)
	} else {
		repo, err = getCurrentRepo()
	}

	if err != nil {
		return nil, err
	}

	return &Client{Repo: repo}, nil
}

func parseRepoArg(arg string) (RepoInfo, error) {
	parts := strings.Split(arg, "/")
	if len(parts) != 2 {
		return RepoInfo{}, fmt.Errorf("invalid repo format: %s. Expected: owner/name", arg)
	}
	return RepoInfo{Owner: parts[0], Name: parts[1]}, nil
}

func getCurrentRepo() (RepoInfo, error) {
	out, err := exec.Command("gh", "repo", "view", "--json", "owner,name").Output()
	if err != nil {
		return RepoInfo{}, fmt.Errorf("could not determine repository. Use --repo flag or run from a git repository")
	}

	var result struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return RepoInfo{}, fmt.Errorf("failed to parse repo info: %w", err)
	}

	return RepoInfo{Owner: result.Owner.Login, Name: result.Name}, nil
}

// ghAPI executes a gh api command and returns the result
func (c *Client) ghAPI(endpoint string, args ...string) ([]byte, error) {
	cmdArgs := []string{"api", endpoint}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("gh", cmdArgs...)
	return cmd.Output()
}

// GetRepo fetches repository settings
func (c *Client) GetRepo() (*RepoData, error) {
	out, err := c.ghAPI(fmt.Sprintf("repos/%s/%s", c.Repo.Owner, c.Repo.Name))
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
func (c *Client) UpdateRepo(settings map[string]interface{}) error {
	jsonData, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = c.ghAPI(
		fmt.Sprintf("repos/%s/%s", c.Repo.Owner, c.Repo.Name),
		"-X", "PATCH",
		"-H", "Accept: application/vnd.github+json",
		"--input", "-",
	)
	if err != nil {
		// Try with field flags instead
		args := []string{"api", fmt.Sprintf("repos/%s/%s", c.Repo.Owner, c.Repo.Name), "-X", "PATCH"}
		for k, v := range settings {
			switch val := v.(type) {
			case string:
				args = append(args, "-f", fmt.Sprintf("%s=%s", k, val))
			case bool:
				args = append(args, "-F", fmt.Sprintf("%s=%t", k, val))
			}
		}
		cmd := exec.Command("gh", args...)
		_, err = cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to update repo: %w, payload: %s", err, string(jsonData))
		}
	}

	return nil
}

// GetLabels fetches repository labels
func (c *Client) GetLabels() ([]LabelData, error) {
	out, err := c.ghAPI(fmt.Sprintf("repos/%s/%s/labels", c.Repo.Owner, c.Repo.Name), "--paginate")
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
func (c *Client) CreateLabel(name, color, description string) error {
	args := []string{
		"api", fmt.Sprintf("repos/%s/%s/labels", c.Repo.Owner, c.Repo.Name),
		"-X", "POST",
		"-f", fmt.Sprintf("name=%s", name),
		"-f", fmt.Sprintf("color=%s", color),
	}
	if description != "" {
		args = append(args, "-f", fmt.Sprintf("description=%s", description))
	}

	cmd := exec.Command("gh", args...)
	_, err := cmd.Output()
	return err
}

// UpdateLabel updates an existing label
func (c *Client) UpdateLabel(oldName, newName, color, description string) error {
	args := []string{
		"api", fmt.Sprintf("repos/%s/%s/labels/%s", c.Repo.Owner, c.Repo.Name, oldName),
		"-X", "PATCH",
		"-f", fmt.Sprintf("new_name=%s", newName),
		"-f", fmt.Sprintf("color=%s", color),
	}
	if description != "" {
		args = append(args, "-f", fmt.Sprintf("description=%s", description))
	}

	cmd := exec.Command("gh", args...)
	_, err := cmd.Output()
	return err
}

// DeleteLabel deletes a label
func (c *Client) DeleteLabel(name string) error {
	cmd := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/labels/%s", c.Repo.Owner, c.Repo.Name, name),
		"-X", "DELETE",
	)
	_, err := cmd.Output()
	return err
}

// SetTopics sets repository topics
func (c *Client) SetTopics(topics []string) error {
	topicsJSON, _ := json.Marshal(topics)
	cmd := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/topics", c.Repo.Owner, c.Repo.Name),
		"-X", "PUT",
		"-H", "Accept: application/vnd.github+json",
		"-f", fmt.Sprintf("names=%s", string(topicsJSON)),
	)
	_, err := cmd.Output()
	return err
}

// GetSecrets fetches repository secret names
func (c *Client) GetSecrets() ([]string, error) {
	out, err := c.ghAPI(fmt.Sprintf("repos/%s/%s/actions/secrets", c.Repo.Owner, c.Repo.Name))
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
func (c *Client) GetVariables() ([]string, error) {
	out, err := c.ghAPI(fmt.Sprintf("repos/%s/%s/actions/variables", c.Repo.Owner, c.Repo.Name))
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
func (c *Client) GetBranchProtection(branch string) (*BranchProtectionData, error) {
	out, err := c.ghAPI(fmt.Sprintf("repos/%s/%s/branches/%s/protection", c.Repo.Owner, c.Repo.Name, branch))
	if err != nil {
		return nil, err
	}

	var data BranchProtectionData
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
