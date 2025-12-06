package workflow

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Job represents a GitHub Actions job
type Job struct {
	Name string `yaml:"name"`
}

// Workflow represents a GitHub Actions workflow file
type Workflow struct {
	Name string         `yaml:"name"`
	Jobs map[string]Job `yaml:"jobs"`
}

// GetCheckNames extracts status check names from workflow files
// The check name is job.name if specified, otherwise the job key
func GetCheckNames(workflowDir string) ([]string, error) {
	if workflowDir == "" {
		workflowDir = ".github/workflows"
	}

	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No workflows directory
		}
		return nil, err
	}

	var checkNames []string
	seen := make(map[string]bool)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		filePath := filepath.Join(workflowDir, name)
		names, err := parseWorkflowFile(filePath)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		for _, n := range names {
			if !seen[n] {
				seen[n] = true
				checkNames = append(checkNames, n)
			}
		}
	}

	return checkNames, nil
}

func parseWorkflowFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, err
	}

	var names []string
	for jobKey, job := range workflow.Jobs {
		// GitHub uses job.name if specified, otherwise the job key
		if job.Name != "" {
			names = append(names, job.Name)
		} else {
			names = append(names, jobKey)
		}
	}

	return names, nil
}

// ValidateStatusChecks validates that all status checks exist in workflows
// Returns a list of unknown check names
func ValidateStatusChecks(statusChecks []string, workflowDir string) ([]string, []string, error) {
	available, err := GetCheckNames(workflowDir)
	if err != nil {
		return nil, nil, err
	}

	if len(available) == 0 {
		return nil, nil, nil // No workflows to validate against
	}

	availableSet := make(map[string]bool)
	for _, name := range available {
		availableSet[name] = true
	}

	var unknown []string
	for _, check := range statusChecks {
		if !availableSet[check] {
			unknown = append(unknown, check)
		}
	}

	return unknown, available, nil
}
