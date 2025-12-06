package diff

import (
	"github.com/myzkey/gh-repo-settings/internal/infra/githubopenapi"
	"github.com/oapi-codegen/nullable"
)

// ptr creates a pointer to a value (test helper)
func ptr[T any](v T) *T {
	return &v
}

// nullStr creates a nullable.Nullable[string] with a value
func nullStr(s string) nullable.Nullable[string] {
	return nullable.NewNullableWithValue(s)
}

// allowedActions creates a pointer to AllowedActions
func allowedActions(s string) *githubopenapi.AllowedActions {
	a := githubopenapi.AllowedActions(s)
	return &a
}

// nullBuildType creates a nullable.Nullable[GithubPageBuildType] with a value
func nullBuildType(s string) nullable.Nullable[githubopenapi.GithubPageBuildType] {
	return nullable.NewNullableWithValue(githubopenapi.GithubPageBuildType(s))
}
