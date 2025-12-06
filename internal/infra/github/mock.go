package github

import (
	"context"

	"github.com/myzkey/gh-repo-settings/internal/infra/githubopenapi"
)

// MockClient is a mock implementation of GitHubClient for testing
type MockClient struct {
	RepoData             *RepoData
	Labels               []LabelData
	BranchProtections    map[string]*BranchProtectionData
	Secrets              []string
	Variables            []VariableData
	ActionsPermissions   *ActionsPermissionsData
	ActionsSelected      *ActionsSelectedData
	ActionsWorkflowPerms *ActionsWorkflowPermissionsData
	PagesData            *PagesData
	Owner                string
	Name                 string

	// Error fields for testing error scenarios
	GetRepoError                       error
	UpdateRepoError                    error
	GetLabelsError                     error
	CreateLabelError                   error
	UpdateLabelError                   error
	DeleteLabelError                   error
	SetTopicsError                     error
	GetBranchProtectionError           error
	UpdateBranchProtectionError        error
	GetSecretsError                    error
	SetSecretError                     error
	DeleteSecretError                  error
	GetVariablesError                  error
	SetVariableError                   error
	DeleteVariableError                error
	GetActionsPermissionsError         error
	UpdateActionsPermissionsError      error
	GetActionsSelectedActionsError     error
	UpdateActionsSelectedActionsError  error
	GetActionsWorkflowPermissionsError error
	UpdateActionsWorkflowPermsError    error
	GetPagesError                      error
	CreatePagesError                   error
	UpdatePagesError                   error

	// Call tracking
	UpdateRepoCalls                 []map[string]interface{}
	SetTopicsCalls                  [][]string
	CreateLabelCalls                []LabelCall
	UpdateLabelCalls                []UpdateLabelCall
	DeleteLabelCalls                []string
	UpdateBranchProtectionCalls     []BranchProtectionCall
	SetSecretCalls                  []SecretCall
	DeleteSecretCalls               []string
	SetVariableCalls                []VariableCall
	DeleteVariableCalls             []string
	UpdateActionsPermissionsCalls   []ActionsPermissionsCall
	UpdateActionsSelectedCalls      []*ActionsSelectedData
	UpdateActionsWorkflowPermsCalls []ActionsWorkflowPermsCall
	CreatePagesCalls                []PagesCall
	UpdatePagesCalls                []PagesCall
}

// SecretCall tracks SetSecret calls
type SecretCall struct {
	Name  string
	Value string
}

// VariableCall tracks SetVariable calls
type VariableCall struct {
	Name  string
	Value string
}

// PagesCall tracks CreatePages and UpdatePages calls
type PagesCall struct {
	BuildType string
	Source    *PagesSourceData
}

// ActionsPermissionsCall tracks UpdateActionsPermissions calls
type ActionsPermissionsCall struct {
	Enabled        bool
	AllowedActions string
}

// ActionsWorkflowPermsCall tracks UpdateActionsWorkflowPermissions calls
type ActionsWorkflowPermsCall struct {
	Permissions string
	CanApprove  bool
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
		Variables:         []VariableData{},
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

// SetSecret records the set secret call
func (m *MockClient) SetSecret(ctx context.Context, name, value string) error {
	if m.SetSecretError != nil {
		return m.SetSecretError
	}
	m.SetSecretCalls = append(m.SetSecretCalls, SecretCall{Name: name, Value: value})
	return nil
}

// DeleteSecret records the delete secret call
func (m *MockClient) DeleteSecret(ctx context.Context, name string) error {
	if m.DeleteSecretError != nil {
		return m.DeleteSecretError
	}
	m.DeleteSecretCalls = append(m.DeleteSecretCalls, name)
	return nil
}

// GetVariables returns mock variables
func (m *MockClient) GetVariables(ctx context.Context) ([]VariableData, error) {
	if m.GetVariablesError != nil {
		return nil, m.GetVariablesError
	}
	return m.Variables, nil
}

// SetVariable records the set variable call
func (m *MockClient) SetVariable(ctx context.Context, name, value string) error {
	if m.SetVariableError != nil {
		return m.SetVariableError
	}
	m.SetVariableCalls = append(m.SetVariableCalls, VariableCall{Name: name, Value: value})
	return nil
}

// DeleteVariable records the delete variable call
func (m *MockClient) DeleteVariable(ctx context.Context, name string) error {
	if m.DeleteVariableError != nil {
		return m.DeleteVariableError
	}
	m.DeleteVariableCalls = append(m.DeleteVariableCalls, name)
	return nil
}

// GetActionsPermissions returns mock actions permissions
func (m *MockClient) GetActionsPermissions(ctx context.Context) (*ActionsPermissionsData, error) {
	if m.GetActionsPermissionsError != nil {
		return nil, m.GetActionsPermissionsError
	}
	if m.ActionsPermissions == nil {
		allowedActions := githubopenapi.AllowedActions("all")
		return &ActionsPermissionsData{Enabled: true, AllowedActions: &allowedActions}, nil
	}
	return m.ActionsPermissions, nil
}

// UpdateActionsPermissions records the update call
func (m *MockClient) UpdateActionsPermissions(ctx context.Context, enabled bool, allowedActions string) error {
	if m.UpdateActionsPermissionsError != nil {
		return m.UpdateActionsPermissionsError
	}
	m.UpdateActionsPermissionsCalls = append(m.UpdateActionsPermissionsCalls, ActionsPermissionsCall{
		Enabled:        enabled,
		AllowedActions: allowedActions,
	})
	return nil
}

// GetActionsSelectedActions returns mock selected actions
func (m *MockClient) GetActionsSelectedActions(ctx context.Context) (*ActionsSelectedData, error) {
	if m.GetActionsSelectedActionsError != nil {
		return nil, m.GetActionsSelectedActionsError
	}
	if m.ActionsSelected == nil {
		return &ActionsSelectedData{}, nil
	}
	return m.ActionsSelected, nil
}

// UpdateActionsSelectedActions records the update call
func (m *MockClient) UpdateActionsSelectedActions(ctx context.Context, settings *ActionsSelectedData) error {
	if m.UpdateActionsSelectedActionsError != nil {
		return m.UpdateActionsSelectedActionsError
	}
	m.UpdateActionsSelectedCalls = append(m.UpdateActionsSelectedCalls, settings)
	return nil
}

// GetActionsWorkflowPermissions returns mock workflow permissions
func (m *MockClient) GetActionsWorkflowPermissions(ctx context.Context) (*ActionsWorkflowPermissionsData, error) {
	if m.GetActionsWorkflowPermissionsError != nil {
		return nil, m.GetActionsWorkflowPermissionsError
	}
	if m.ActionsWorkflowPerms == nil {
		return &ActionsWorkflowPermissionsData{DefaultWorkflowPermissions: githubopenapi.ActionsDefaultWorkflowPermissions("read"), CanApprovePullRequestReviews: false}, nil
	}
	return m.ActionsWorkflowPerms, nil
}

// UpdateActionsWorkflowPermissions records the update call
func (m *MockClient) UpdateActionsWorkflowPermissions(ctx context.Context, permissions string, canApprove bool) error {
	if m.UpdateActionsWorkflowPermsError != nil {
		return m.UpdateActionsWorkflowPermsError
	}
	m.UpdateActionsWorkflowPermsCalls = append(m.UpdateActionsWorkflowPermsCalls, ActionsWorkflowPermsCall{
		Permissions: permissions,
		CanApprove:  canApprove,
	})
	return nil
}

// GetPages returns mock pages data
func (m *MockClient) GetPages(ctx context.Context) (*PagesData, error) {
	if m.GetPagesError != nil {
		return nil, m.GetPagesError
	}
	return m.PagesData, nil
}

// CreatePages records the create call
func (m *MockClient) CreatePages(ctx context.Context, buildType string, source *PagesSourceData) error {
	if m.CreatePagesError != nil {
		return m.CreatePagesError
	}
	m.CreatePagesCalls = append(m.CreatePagesCalls, PagesCall{
		BuildType: buildType,
		Source:    source,
	})
	return nil
}

// UpdatePages records the update call
func (m *MockClient) UpdatePages(ctx context.Context, buildType string, source *PagesSourceData) error {
	if m.UpdatePagesError != nil {
		return m.UpdatePagesError
	}
	m.UpdatePagesCalls = append(m.UpdatePagesCalls, PagesCall{
		BuildType: buildType,
		Source:    source,
	})
	return nil
}

// Ensure MockClient implements GitHubClient
var _ GitHubClient = (*MockClient)(nil)
