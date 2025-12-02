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

func TestParseHTTPStatus(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   int
	}{
		{
			name:   "404 not found",
			stderr: "gh: Not Found (HTTP 404)",
			want:   404,
		},
		{
			name:   "403 forbidden",
			stderr: "gh: Resource not accessible by integration (HTTP 403)",
			want:   403,
		},
		{
			name:   "401 unauthorized",
			stderr: "gh: Bad credentials (HTTP 401)",
			want:   401,
		},
		{
			name:   "422 unprocessable",
			stderr: "gh: Validation Failed (HTTP 422)",
			want:   422,
		},
		{
			name:   "500 server error",
			stderr: "gh: Internal Server Error (HTTP 500)",
			want:   500,
		},
		{
			name:   "no http status",
			stderr: "some other error message",
			want:   0,
		},
		{
			name:   "empty stderr",
			stderr: "",
			want:   0,
		},
		{
			name:   "multiline with status",
			stderr: "some error\ngh: Not Found (HTTP 404)\nmore info",
			want:   404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHTTPStatus(tt.stderr)
			if got != tt.want {
				t.Errorf("parseHTTPStatus(%q) = %d, want %d", tt.stderr, got, tt.want)
			}
		})
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

func TestBranchProtectionSettingsAllFields(t *testing.T) {
	reviews := 1
	dismiss := true
	codeOwner := true
	statusChecks := true
	strict := true
	linear := true
	forcePush := true
	deletions := true
	signed := true
	admins := true

	settings := &BranchProtectionSettings{
		RequiredReviews:         &reviews,
		DismissStaleReviews:     &dismiss,
		RequireCodeOwnerReviews: &codeOwner,
		RequireStatusChecks:     &statusChecks,
		StatusChecks:            []string{"test", "lint"},
		StrictStatusChecks:      &strict,
		RequireLinearHistory:    &linear,
		AllowForcePushes:        &forcePush,
		AllowDeletions:          &deletions,
		RequireSignedCommits:    &signed,
		EnforceAdmins:           &admins,
	}

	if *settings.DismissStaleReviews != true {
		t.Error("DismissStaleReviews should be true")
	}
	if *settings.RequireCodeOwnerReviews != true {
		t.Error("RequireCodeOwnerReviews should be true")
	}
	if *settings.RequireStatusChecks != true {
		t.Error("RequireStatusChecks should be true")
	}
	if *settings.RequireLinearHistory != true {
		t.Error("RequireLinearHistory should be true")
	}
	if *settings.AllowForcePushes != true {
		t.Error("AllowForcePushes should be true")
	}
	if *settings.AllowDeletions != true {
		t.Error("AllowDeletions should be true")
	}
	if *settings.RequireSignedCommits != true {
		t.Error("RequireSignedCommits should be true")
	}
}

func TestRepoData(t *testing.T) {
	desc := "Test repo"
	homepage := "https://example.com"

	data := &RepoData{
		Description:         &desc,
		Homepage:            &homepage,
		Visibility:          "public",
		AllowMergeCommit:    true,
		AllowRebaseMerge:    true,
		AllowSquashMerge:    true,
		DeleteBranchOnMerge: true,
		AllowUpdateBranch:   true,
		Topics:              []string{"go", "cli"},
	}

	if *data.Description != "Test repo" {
		t.Errorf("expected description 'Test repo', got %q", *data.Description)
	}
	if data.Visibility != "public" {
		t.Errorf("expected visibility 'public', got %q", data.Visibility)
	}
	if len(data.Topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(data.Topics))
	}
}

func TestLabelData(t *testing.T) {
	label := LabelData{
		Name:        "bug",
		Color:       "d73a4a",
		Description: "Something isn't working",
	}

	if label.Name != "bug" {
		t.Errorf("expected name 'bug', got %q", label.Name)
	}
	if label.Color != "d73a4a" {
		t.Errorf("expected color 'd73a4a', got %q", label.Color)
	}
}

func TestVariableData(t *testing.T) {
	variable := VariableData{
		Name:  "NODE_ENV",
		Value: "production",
	}

	if variable.Name != "NODE_ENV" {
		t.Errorf("expected name 'NODE_ENV', got %q", variable.Name)
	}
	if variable.Value != "production" {
		t.Errorf("expected value 'production', got %q", variable.Value)
	}
}

func TestActionsPermissionsData(t *testing.T) {
	data := ActionsPermissionsData{
		Enabled:        true,
		AllowedActions: "selected",
	}

	if !data.Enabled {
		t.Error("expected Enabled to be true")
	}
	if data.AllowedActions != "selected" {
		t.Errorf("expected AllowedActions 'selected', got %q", data.AllowedActions)
	}
}

func TestActionsSelectedData(t *testing.T) {
	data := ActionsSelectedData{
		GithubOwnedAllowed: true,
		VerifiedAllowed:    true,
		PatternsAllowed:    []string{"actions/*", "github/*"},
	}

	if !data.GithubOwnedAllowed {
		t.Error("expected GithubOwnedAllowed to be true")
	}
	if !data.VerifiedAllowed {
		t.Error("expected VerifiedAllowed to be true")
	}
	if len(data.PatternsAllowed) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(data.PatternsAllowed))
	}
}

func TestActionsWorkflowPermissionsData(t *testing.T) {
	data := ActionsWorkflowPermissionsData{
		DefaultWorkflowPermissions:   "read",
		CanApprovePullRequestReviews: false,
	}

	if data.DefaultWorkflowPermissions != "read" {
		t.Errorf("expected DefaultWorkflowPermissions 'read', got %q", data.DefaultWorkflowPermissions)
	}
	if data.CanApprovePullRequestReviews {
		t.Error("expected CanApprovePullRequestReviews to be false")
	}
}

func TestPagesData(t *testing.T) {
	data := PagesData{
		BuildType: "workflow",
		Source: &PagesSourceData{
			Branch: "main",
			Path:   "/docs",
		},
	}

	if data.BuildType != "workflow" {
		t.Errorf("expected BuildType 'workflow', got %q", data.BuildType)
	}
	if data.Source.Branch != "main" {
		t.Errorf("expected Source.Branch 'main', got %q", data.Source.Branch)
	}
	if data.Source.Path != "/docs" {
		t.Errorf("expected Source.Path '/docs', got %q", data.Source.Path)
	}
}

func TestPagesSourceData(t *testing.T) {
	source := PagesSourceData{
		Branch: "gh-pages",
		Path:   "/",
	}

	if source.Branch != "gh-pages" {
		t.Errorf("expected Branch 'gh-pages', got %q", source.Branch)
	}
	if source.Path != "/" {
		t.Errorf("expected Path '/', got %q", source.Path)
	}
}

func TestBranchProtectionData(t *testing.T) {
	data := BranchProtectionData{
		RequiredPullRequestReviews: &struct {
			RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
			DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
			RequireCodeOwnerReviews      bool `json:"require_code_owner_reviews"`
		}{
			RequiredApprovingReviewCount: 2,
			DismissStaleReviews:          true,
			RequireCodeOwnerReviews:      true,
		},
		RequiredStatusChecks: &struct {
			Strict   bool     `json:"strict"`
			Contexts []string `json:"contexts"`
		}{
			Strict:   true,
			Contexts: []string{"ci/test", "ci/lint"},
		},
		EnforceAdmins: &struct {
			Enabled bool `json:"enabled"`
		}{
			Enabled: true,
		},
		RequiredLinearHistory: &struct {
			Enabled bool `json:"enabled"`
		}{
			Enabled: true,
		},
		AllowForcePushes: &struct {
			Enabled bool `json:"enabled"`
		}{
			Enabled: false,
		},
		AllowDeletions: &struct {
			Enabled bool `json:"enabled"`
		}{
			Enabled: false,
		},
		RequiredSignatures: &struct {
			Enabled bool `json:"enabled"`
		}{
			Enabled: true,
		},
	}

	if data.RequiredPullRequestReviews.RequiredApprovingReviewCount != 2 {
		t.Errorf("expected 2 required reviews, got %d", data.RequiredPullRequestReviews.RequiredApprovingReviewCount)
	}
	if !data.RequiredPullRequestReviews.DismissStaleReviews {
		t.Error("expected DismissStaleReviews to be true")
	}
	if !data.RequiredStatusChecks.Strict {
		t.Error("expected Strict status checks to be true")
	}
	if len(data.RequiredStatusChecks.Contexts) != 2 {
		t.Errorf("expected 2 status check contexts, got %d", len(data.RequiredStatusChecks.Contexts))
	}
	if !data.EnforceAdmins.Enabled {
		t.Error("expected EnforceAdmins to be enabled")
	}
}

func TestBranchProtectionDataNilFields(t *testing.T) {
	// Test with nil optional fields
	data := BranchProtectionData{}

	if data.RequiredPullRequestReviews != nil {
		t.Error("expected RequiredPullRequestReviews to be nil")
	}
	if data.RequiredStatusChecks != nil {
		t.Error("expected RequiredStatusChecks to be nil")
	}
	if data.EnforceAdmins != nil {
		t.Error("expected EnforceAdmins to be nil")
	}
}

func TestRepoInfo(t *testing.T) {
	info := RepoInfo{
		Owner: "test-owner",
		Name:  "test-repo",
	}

	if info.Owner != "test-owner" {
		t.Errorf("expected Owner 'test-owner', got %q", info.Owner)
	}
	if info.Name != "test-repo" {
		t.Errorf("expected Name 'test-repo', got %q", info.Name)
	}
}

func TestParseRepoArgEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    RepoInfo
		wantErr bool
	}{
		{
			name:    "valid with numbers",
			arg:     "owner123/repo456",
			want:    RepoInfo{Owner: "owner123", Name: "repo456"},
			wantErr: false,
		},
		{
			name:    "valid with underscores",
			arg:     "my_org/my_repo",
			want:    RepoInfo{Owner: "my_org", Name: "my_repo"},
			wantErr: false,
		},
		{
			name:    "empty owner",
			arg:     "/repo",
			want:    RepoInfo{Owner: "", Name: "repo"},
			wantErr: false, // parseRepoArg doesn't validate empty parts
		},
		{
			name:    "empty repo",
			arg:     "owner/",
			want:    RepoInfo{Owner: "owner", Name: ""},
			wantErr: false, // parseRepoArg doesn't validate empty parts
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

func TestHTTPStatusRegex(t *testing.T) {
	// Test that the regex is properly compiled and works
	testCases := []struct {
		input    string
		expected bool
	}{
		{"HTTP 200", true},
		{"HTTP 404", true},
		{"HTTP 500", true},
		{"http 200", false}, // case sensitive
		{"HTTP", false},
		{"HTTP abc", false},
		{"", false},
	}

	for _, tc := range testCases {
		matches := httpStatusRegex.FindStringSubmatch(tc.input)
		hasMatch := len(matches) >= 2
		if hasMatch != tc.expected {
			t.Errorf("httpStatusRegex.FindStringSubmatch(%q) match = %v, want %v", tc.input, hasMatch, tc.expected)
		}
	}
}
