package github

import (
	"testing"
)

func TestParseRepoArg(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    RepoInfo
		wantErr bool
	}{
		{
			name:    "valid repo",
			arg:     "owner/repo",
			want:    RepoInfo{Owner: "owner", Name: "repo"},
			wantErr: false,
		},
		{
			name:    "repo with dashes",
			arg:     "my-org/my-repo",
			want:    RepoInfo{Owner: "my-org", Name: "my-repo"},
			wantErr: false,
		},
		{
			name:    "invalid - no slash",
			arg:     "invalid",
			wantErr: true,
		},
		{
			name:    "invalid - too many slashes",
			arg:     "a/b/c",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			arg:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRepoArg(tt.arg)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got.Owner != tt.want.Owner || got.Name != tt.want.Name {
				t.Errorf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}

func TestClientRepoOwnerAndName(t *testing.T) {
	client := &Client{
		Repo: RepoInfo{
			Owner: "test-owner",
			Name:  "test-repo",
		},
	}

	if client.RepoOwner() != "test-owner" {
		t.Errorf("expected owner 'test-owner', got '%s'", client.RepoOwner())
	}

	if client.RepoName() != "test-repo" {
		t.Errorf("expected name 'test-repo', got '%s'", client.RepoName())
	}
}

func TestMockClientImplementsInterface(t *testing.T) {
	// This test verifies that MockClient implements GitHubClient
	var _ GitHubClient = (*MockClient)(nil)
	var _ GitHubClient = (*Client)(nil)
}

func TestMockClient(t *testing.T) {
	mock := NewMockClient()

	// Test default values
	if mock.RepoOwner() != "test-owner" {
		t.Errorf("expected default owner 'test-owner', got '%s'", mock.RepoOwner())
	}

	if mock.RepoName() != "test-repo" {
		t.Errorf("expected default name 'test-repo', got '%s'", mock.RepoName())
	}

	// Test custom values
	mock.Owner = "custom-owner"
	mock.Name = "custom-repo"

	if mock.RepoOwner() != "custom-owner" {
		t.Errorf("expected custom owner 'custom-owner', got '%s'", mock.RepoOwner())
	}
}

func TestBranchProtectionSettings(t *testing.T) {
	reviews := 2
	strict := true
	enforceAdmins := true

	settings := &BranchProtectionSettings{
		RequiredReviews:    &reviews,
		StrictStatusChecks: &strict,
		EnforceAdmins:      &enforceAdmins,
		StatusChecks:       []string{"ci/build", "ci/test"},
	}

	if *settings.RequiredReviews != 2 {
		t.Errorf("expected 2 required reviews")
	}

	if len(settings.StatusChecks) != 2 {
		t.Errorf("expected 2 status checks")
	}
}
