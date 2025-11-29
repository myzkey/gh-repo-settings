package github

import (
	"context"
)

// MockClient is a mock implementation of GitHubClient for testing
type MockClient struct {
	RepoData              *RepoData
	Labels                []LabelData
	BranchProtections     map[string]*BranchProtectionData
	Secrets               []string
	Variables             []string
	Owner                 string
	Name                  string

	// Error fields for testing error scenarios
	GetRepoError              error
	UpdateRepoError           error
	GetLabelsError            error
	CreateLabelError          error
	UpdateLabelError          error
	DeleteLabelError          error
	SetTopicsError            error
	GetBranchProtectionError  error
	UpdateBranchProtectionError error
	GetSecretsError           error
	GetVariablesError         error

	// Call tracking
	UpdateRepoCalls           []map[string]interface{}
	SetTopicsCalls            [][]string
	CreateLabelCalls          []LabelCall
	UpdateLabelCalls          []UpdateLabelCall
	DeleteLabelCalls          []string
	UpdateBranchProtectionCalls []BranchProtectionCall
}

// LabelCall tracks CreateLabel calls
type LabelCall struct {
	Name        string
	Color       string
	Description string
}

// UpdateLabelCall tracks UpdateLabel calls
type UpdateLabelCall struct {
	OldName     string
	NewName     string
	Color       string
	Description string
}

// BranchProtectionCall tracks UpdateBranchProtection calls
type BranchProtectionCall struct {
	Branch   string
	Settings *BranchProtectionSettings
}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{
		RepoData:          &RepoData{},
		Labels:            []LabelData{},
		BranchProtections: make(map[string]*BranchProtectionData),
		Secrets:           []string{},
		Variables:         []string{},
		Owner:             "test-owner",
		Name:              "test-repo",
	}
}

// RepoOwner returns the mock owner
func (m *MockClient) RepoOwner() string {
	return m.Owner
}

// RepoName returns the mock repo name
func (m *MockClient) RepoName() string {
	return m.Name
}

// GetRepo returns mock repo data
func (m *MockClient) GetRepo(ctx context.Context) (*RepoData, error) {
	if m.GetRepoError != nil {
		return nil, m.GetRepoError
	}
	return m.RepoData, nil
}

// UpdateRepo records the update call
func (m *MockClient) UpdateRepo(ctx context.Context, settings map[string]interface{}) error {
	if m.UpdateRepoError != nil {
		return m.UpdateRepoError
	}
	m.UpdateRepoCalls = append(m.UpdateRepoCalls, settings)
	return nil
}

// SetTopics records the topics call
func (m *MockClient) SetTopics(ctx context.Context, topics []string) error {
	if m.SetTopicsError != nil {
		return m.SetTopicsError
	}
	m.SetTopicsCalls = append(m.SetTopicsCalls, topics)
	return nil
}

// GetLabels returns mock labels
func (m *MockClient) GetLabels(ctx context.Context) ([]LabelData, error) {
	if m.GetLabelsError != nil {
		return nil, m.GetLabelsError
	}
	return m.Labels, nil
}

// CreateLabel records the create call
func (m *MockClient) CreateLabel(ctx context.Context, name, color, description string) error {
	if m.CreateLabelError != nil {
		return m.CreateLabelError
	}
	m.CreateLabelCalls = append(m.CreateLabelCalls, LabelCall{
		Name:        name,
		Color:       color,
		Description: description,
	})
	return nil
}

// UpdateLabel records the update call
func (m *MockClient) UpdateLabel(ctx context.Context, oldName, newName, color, description string) error {
	if m.UpdateLabelError != nil {
		return m.UpdateLabelError
	}
	m.UpdateLabelCalls = append(m.UpdateLabelCalls, UpdateLabelCall{
		OldName:     oldName,
		NewName:     newName,
		Color:       color,
		Description: description,
	})
	return nil
}

// DeleteLabel records the delete call
func (m *MockClient) DeleteLabel(ctx context.Context, name string) error {
	if m.DeleteLabelError != nil {
		return m.DeleteLabelError
	}
	m.DeleteLabelCalls = append(m.DeleteLabelCalls, name)
	return nil
}

// GetBranchProtection returns mock branch protection
func (m *MockClient) GetBranchProtection(ctx context.Context, branch string) (*BranchProtectionData, error) {
	if m.GetBranchProtectionError != nil {
		return nil, m.GetBranchProtectionError
	}
	if bp, ok := m.BranchProtections[branch]; ok {
		return bp, nil
	}
	return nil, nil
}

// UpdateBranchProtection records the update call
func (m *MockClient) UpdateBranchProtection(ctx context.Context, branch string, settings *BranchProtectionSettings) error {
	if m.UpdateBranchProtectionError != nil {
		return m.UpdateBranchProtectionError
	}
	m.UpdateBranchProtectionCalls = append(m.UpdateBranchProtectionCalls, BranchProtectionCall{
		Branch:   branch,
		Settings: settings,
	})
	return nil
}

// GetSecrets returns mock secrets
func (m *MockClient) GetSecrets(ctx context.Context) ([]string, error) {
	if m.GetSecretsError != nil {
		return nil, m.GetSecretsError
	}
	return m.Secrets, nil
}

// GetVariables returns mock variables
func (m *MockClient) GetVariables(ctx context.Context) ([]string, error) {
	if m.GetVariablesError != nil {
		return nil, m.GetVariablesError
	}
	return m.Variables, nil
}

// Ensure MockClient implements GitHubClient
var _ GitHubClient = (*MockClient)(nil)
