package comparator

import (
	"github.com/myzkey/gh-repo-settings/internal/infra/githubopenapi"
	"github.com/oapi-codegen/nullable"
)

func ptr[T any](v T) *T {
	return &v
}

func nullStr(s string) nullable.Nullable[string] {
	return nullable.NewNullableWithValue(s)
}

func nullBuildType(s string) nullable.Nullable[githubopenapi.GithubPageBuildType] {
	return nullable.NewNullableWithValue(githubopenapi.GithubPageBuildType(s))
}

func allowedActions(s string) *githubopenapi.AllowedActions {
	a := githubopenapi.AllowedActions(s)
	return &a
}
