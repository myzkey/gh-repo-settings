package github

import (
	"testing"
)

func TestRepoPath(t *testing.T) {
	client := &Client{
		Repo: RepoInfo{
			Owner: "owner",
			Name:  "repo",
		},
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "empty path returns base repo path",
			path:     "",
			expected: "repos/owner/repo",
		},
		{
			name:     "labels path",
			path:     "labels",
			expected: "repos/owner/repo/labels",
		},
		{
			name:     "nested path",
			path:     "branches/main/protection",
			expected: "repos/owner/repo/branches/main/protection",
		},
		{
			name:     "actions/secrets path",
			path:     "actions/secrets",
			expected: "repos/owner/repo/actions/secrets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.repoPath(tt.path)
			if result != tt.expected {
				t.Errorf("repoPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestJsonHeaders(t *testing.T) {
	headers := jsonHeaders()

	if len(headers) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(headers))
	}

	if headers[0] != "-H" {
		t.Errorf("headers[0] = %q, want %q", headers[0], "-H")
	}

	if headers[1] != "Accept: application/vnd.github+json" {
		t.Errorf("headers[1] = %q, want %q", headers[1], "Accept: application/vnd.github+json")
	}
}

func TestHTTPMethodConstants(t *testing.T) {
	tests := []struct {
		method   httpMethod
		expected string
	}{
		{httpGet, "GET"},
		{httpPost, "POST"},
		{httpPut, "PUT"},
		{httpPatch, "PATCH"},
		{httpDelete, "DELETE"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.method) != tt.expected {
				t.Errorf("httpMethod = %q, want %q", string(tt.method), tt.expected)
			}
		})
	}
}

func TestCallAPI_GetWithBodyReturnsError(t *testing.T) {
	client := &Client{
		Repo: RepoInfo{Owner: "owner", Name: "repo"},
	}

	_, err := client.callAPI(t.Context(), httpGet, "test/endpoint", []byte(`{"test": true}`))
	if err == nil {
		t.Fatal("expected error for GET request with body, got nil")
	}

	expectedMsg := "GET request must not have body"
	if err.Error() != expectedMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestLabelPath(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		expected string
	}{
		{
			name:     "simple label",
			label:    "bug",
			expected: "labels/bug",
		},
		{
			name:     "label with space",
			label:    "help wanted",
			expected: "labels/help%20wanted",
		},
		{
			name:     "Japanese label",
			label:    "バグ",
			expected: "labels/%E3%83%90%E3%82%B0",
		},
		{
			name:     "label with special characters",
			label:    "bug/critical",
			expected: "labels/bug%2Fcritical",
		},
		{
			name:     "label with colon (not encoded per RFC 3986)",
			label:    "priority:high",
			expected: "labels/priority:high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := labelPath(tt.label)
			if result != tt.expected {
				t.Errorf("labelPath(%q) = %q, want %q", tt.label, result, tt.expected)
			}
		})
	}
}

func TestBranchPath(t *testing.T) {
	client := &Client{
		Repo: RepoInfo{
			Owner: "owner",
			Name:  "repo",
		},
	}

	tests := []struct {
		name     string
		branch   string
		suffix   string
		expected string
	}{
		{
			name:     "simple branch with suffix",
			branch:   "main",
			suffix:   "protection",
			expected: "repos/owner/repo/branches/main/protection",
		},
		{
			name:     "simple branch without suffix",
			branch:   "main",
			suffix:   "",
			expected: "repos/owner/repo/branches/main",
		},
		{
			name:     "branch with slash is URL encoded",
			branch:   "feature/foo",
			suffix:   "protection",
			expected: "repos/owner/repo/branches/feature%2Ffoo/protection",
		},
		{
			name:     "branch with multiple slashes",
			branch:   "feature/user/login",
			suffix:   "protection",
			expected: "repos/owner/repo/branches/feature%2Fuser%2Flogin/protection",
		},
		{
			name:     "branch with special characters",
			branch:   "release-v1.0",
			suffix:   "protection",
			expected: "repos/owner/repo/branches/release-v1.0/protection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.branchPath(tt.branch, tt.suffix)
			if result != tt.expected {
				t.Errorf("branchPath(%q, %q) = %q, want %q", tt.branch, tt.suffix, result, tt.expected)
			}
		})
	}
}

func TestSecretPath(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		expected string
	}{
		{
			name:     "simple secret name",
			secret:   "MY_SECRET",
			expected: "actions/secrets/MY_SECRET",
		},
		{
			name:     "secret with special characters",
			secret:   "MY/SECRET",
			expected: "actions/secrets/MY%2FSECRET",
		},
		{
			name:     "secret with space",
			secret:   "MY SECRET",
			expected: "actions/secrets/MY%20SECRET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := secretPath(tt.secret)
			if result != tt.expected {
				t.Errorf("secretPath(%q) = %q, want %q", tt.secret, result, tt.expected)
			}
		})
	}
}

func TestVariablePath(t *testing.T) {
	tests := []struct {
		name     string
		variable string
		expected string
	}{
		{
			name:     "simple variable name",
			variable: "MY_VAR",
			expected: "actions/variables/MY_VAR",
		},
		{
			name:     "variable with special characters",
			variable: "MY/VAR",
			expected: "actions/variables/MY%2FVAR",
		},
		{
			name:     "variable with space",
			variable: "MY VAR",
			expected: "actions/variables/MY%20VAR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := variablePath(tt.variable)
			if result != tt.expected {
				t.Errorf("variablePath(%q) = %q, want %q", tt.variable, result, tt.expected)
			}
		})
	}
}
