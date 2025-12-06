package model

import (
	"testing"
)

func ptr[T any](v T) *T {
	return &v
}

func TestToStringSet(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  map[string]bool
	}{
		{
			name:  "empty slice",
			input: []string{},
			want:  map[string]bool{},
		},
		{
			name:  "single item",
			input: []string{"a"},
			want:  map[string]bool{"a": true},
		},
		{
			name:  "multiple items",
			input: []string{"a", "b", "c"},
			want:  map[string]bool{"a": true, "b": true, "c": true},
		},
		{
			name:  "duplicates",
			input: []string{"a", "b", "a"},
			want:  map[string]bool{"a": true, "b": true},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToStringSet(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("ToStringSet() length = %d, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("ToStringSet()[%s] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestJoinParts(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "empty",
			parts:    []string{},
			expected: "",
		},
		{
			name:     "single",
			parts:    []string{"a=1"},
			expected: "a=1",
		},
		{
			name:     "multiple",
			parts:    []string{"a=1", "b=2", "c=3"},
			expected: "a=1, b=2, c=3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinParts(tt.parts)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPtrStringEqual(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *string
		current  *string
		expected bool
	}{
		{
			name:     "cfg nil (not specified) returns true",
			cfg:      nil,
			current:  ptr("test"),
			expected: true,
		},
		{
			name:     "cfg not nil current nil",
			cfg:      ptr("test"),
			current:  nil,
			expected: false,
		},
		{
			name:     "both equal",
			cfg:      ptr("test"),
			current:  ptr("test"),
			expected: true,
		},
		{
			name:     "both not equal",
			cfg:      ptr("test1"),
			current:  ptr("test2"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PtrStringEqual(tt.cfg, tt.current)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPtrVal(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: "",
		},
		{
			name:     "not nil",
			input:    ptr("test"),
			expected: "test",
		},
		{
			name:     "empty string",
			input:    ptr(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PtrVal(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStringSliceEqualIgnoreOrder(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{
			name:     "both empty",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "same order",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "different order",
			a:        []string{"c", "a", "b"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "different length",
			a:        []string{"a", "b"},
			b:        []string{"a", "b", "c"},
			expected: false,
		},
		{
			name:     "different content",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "d"},
			expected: false,
		},
		{
			name:     "duplicates same",
			a:        []string{"a", "a", "b"},
			b:        []string{"a", "b", "a"},
			expected: true,
		},
		{
			name:     "duplicates different count",
			a:        []string{"a", "a", "b"},
			b:        []string{"a", "b", "b"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringSliceEqualIgnoreOrder(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("StringSliceEqualIgnoreOrder(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
