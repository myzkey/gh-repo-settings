package model

import "github.com/oapi-codegen/nullable"

// ToStringSet converts a slice of strings to a set (map[string]bool)
func ToStringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}

// StringSliceEqualIgnoreOrder compares two string slices ignoring order
func StringSliceEqualIgnoreOrder(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aMap := make(map[string]int)
	for _, v := range a {
		aMap[v]++
	}
	for _, v := range b {
		aMap[v]--
		if aMap[v] < 0 {
			return false
		}
	}
	return true
}

// PtrVal returns the string value from a *string, defaulting to empty string if nil
func PtrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// NullableStringEqual compares a *string (from config) with a nullable.Nullable[string] (from API)
func NullableStringEqual(cfg *string, current nullable.Nullable[string]) bool {
	if cfg == nil {
		return true // config not specified, no comparison needed
	}
	if !current.IsSpecified() {
		return false
	}
	if current.IsNull() {
		return *cfg == ""
	}
	return *cfg == current.MustGet()
}

// NullableStringVal returns the string value from a nullable.Nullable[string]
func NullableStringVal(n nullable.Nullable[string]) string {
	if !n.IsSpecified() || n.IsNull() {
		return ""
	}
	return n.MustGet()
}

// PtrBoolEqual compares *bool values, returning true if they're equal or if cfg is nil
func PtrBoolEqual(cfg, current *bool) bool {
	if cfg == nil {
		return true // config not specified
	}
	if current == nil {
		return false
	}
	return *cfg == *current
}

// PtrBoolVal returns the bool value from a *bool, defaulting to false if nil
func PtrBoolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// PtrStringEqual compares *string values, returning true if they're equal or if cfg is nil
func PtrStringEqual(cfg, current *string) bool {
	if cfg == nil {
		return true // config not specified
	}
	if current == nil {
		return false
	}
	return *cfg == *current
}

// JoinParts joins string parts with ", " separator
func JoinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
