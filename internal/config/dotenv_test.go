package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "dotenv-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name       string
		content    string
		wantValues map[string]string
		wantErr    bool
	}{
		{
			name:    "valid key=value pairs",
			content: "KEY1=value1\nKEY2=value2\n",
			wantValues: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name:    "with comments",
			content: "# This is a comment\nKEY1=value1\n# Another comment\nKEY2=value2\n",
			wantValues: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name:    "with empty lines",
			content: "KEY1=value1\n\n\nKEY2=value2\n",
			wantValues: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name:    "with double quotes",
			content: "KEY1=\"quoted value\"\n",
			wantValues: map[string]string{
				"KEY1": "quoted value",
			},
		},
		{
			name:    "with single quotes",
			content: "KEY1='quoted value'\n",
			wantValues: map[string]string{
				"KEY1": "quoted value",
			},
		},
		{
			name:    "value with equals sign",
			content: "KEY1=value=with=equals\n",
			wantValues: map[string]string{
				"KEY1": "value=with=equals",
			},
		},
		{
			name:    "empty value",
			content: "KEY1=\n",
			wantValues: map[string]string{
				"KEY1": "",
			},
		},
		{
			name:    "malformed line without equals (skipped)",
			content: "MALFORMED_LINE\nKEY1=value1\n",
			wantValues: map[string]string{
				"KEY1": "value1",
			},
		},
		{
			name:       "empty file",
			content:    "",
			wantValues: map[string]string{},
		},
		{
			name:       "only comments",
			content:    "# comment1\n# comment2\n",
			wantValues: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create .env file
			envPath := filepath.Join(tmpDir, ".env")
			if err := os.WriteFile(envPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write .env file: %v", err)
			}

			got, err := LoadDotEnv(tmpDir)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got.Values) != len(tt.wantValues) {
				t.Errorf("got %d values, want %d", len(got.Values), len(tt.wantValues))
			}

			for k, v := range tt.wantValues {
				if got.Values[k] != v {
					t.Errorf("key %q: got %q, want %q", k, got.Values[k], v)
				}
			}
		})
	}
}

func TestLoadDotEnvNonExistent(t *testing.T) {
	got, err := LoadDotEnv("/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error for non-existent file: %v", err)
	}
	if len(got.Values) != 0 {
		t.Errorf("expected empty values for non-existent file, got %d values", len(got.Values))
	}
}

func TestDotEnvValuesGetVariable(t *testing.T) {
	d := &DotEnvValues{
		Values: map[string]string{
			"KEY1": "envValue",
		},
	}

	tests := []struct {
		name        string
		key         string
		yamlDefault string
		want        string
	}{
		{
			name:        "key exists in env",
			key:         "KEY1",
			yamlDefault: "defaultValue",
			want:        "envValue",
		},
		{
			name:        "key not in env, use default",
			key:         "KEY2",
			yamlDefault: "defaultValue",
			want:        "defaultValue",
		},
		{
			name:        "key not in env, empty default",
			key:         "KEY3",
			yamlDefault: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.GetVariable(tt.key, tt.yamlDefault)
			if got != tt.want {
				t.Errorf("GetVariable(%q, %q) = %q, want %q", tt.key, tt.yamlDefault, got, tt.want)
			}
		})
	}
}

func TestDotEnvValuesGetSecret(t *testing.T) {
	d := &DotEnvValues{
		Values: map[string]string{
			"SECRET1": "secretValue",
		},
	}

	tests := []struct {
		name   string
		key    string
		want   string
		wantOk bool
	}{
		{
			name:   "secret exists",
			key:    "SECRET1",
			want:   "secretValue",
			wantOk: true,
		},
		{
			name:   "secret not found",
			key:    "SECRET2",
			want:   "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := d.GetSecret(tt.key)
			if got != tt.want || ok != tt.wantOk {
				t.Errorf("GetSecret(%q) = (%q, %v), want (%q, %v)", tt.key, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}

func TestDotEnvValuesHasValue(t *testing.T) {
	d := &DotEnvValues{
		Values: map[string]string{
			"KEY1": "value",
		},
	}

	if !d.HasValue("KEY1") {
		t.Error("HasValue(KEY1) = false, want true")
	}
	if d.HasValue("KEY2") {
		t.Error("HasValue(KEY2) = true, want false")
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `"quoted"`, want: "quoted"},
		{input: `'quoted'`, want: "quoted"},
		{input: `unquoted`, want: "unquoted"},
		{input: `"`, want: `"`},
		{input: `""`, want: ""},
		{input: `''`, want: ""},
		{input: `"mismatched'`, want: `"mismatched'`},
		{input: ``, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := unquote(tt.input)
			if got != tt.want {
				t.Errorf("unquote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDotEnvValuesMerge(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]string
		other    *DotEnvValues
		expected map[string]string
	}{
		{
			name: "merge non-overlapping keys",
			base: map[string]string{
				"KEY1": "value1",
			},
			other: &DotEnvValues{
				Values: map[string]string{
					"KEY2": "value2",
				},
			},
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "existing values not overwritten",
			base: map[string]string{
				"KEY1": "original",
			},
			other: &DotEnvValues{
				Values: map[string]string{
					"KEY1": "new",
				},
			},
			expected: map[string]string{
				"KEY1": "original",
			},
		},
		{
			name: "merge with nil other",
			base: map[string]string{
				"KEY1": "value1",
			},
			other: nil,
			expected: map[string]string{
				"KEY1": "value1",
			},
		},
		{
			name: "merge with empty other",
			base: map[string]string{
				"KEY1": "value1",
			},
			other: &DotEnvValues{
				Values: map[string]string{},
			},
			expected: map[string]string{
				"KEY1": "value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DotEnvValues{Values: tt.base}
			d.Merge(tt.other)

			if len(d.Values) != len(tt.expected) {
				t.Errorf("after Merge, got %d values, want %d", len(d.Values), len(tt.expected))
			}

			for k, v := range tt.expected {
				if d.Values[k] != v {
					t.Errorf("after Merge, key %q = %q, want %q", k, d.Values[k], v)
				}
			}
		})
	}
}

func TestResolveDotEnvPath(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "resolve-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .github directory
	githubDir := filepath.Join(tmpDir, ".github")
	if err := os.MkdirAll(githubDir, 0755); err != nil {
		t.Fatalf("failed to create .github dir: %v", err)
	}

	// Create a config file
	configFile := filepath.Join(githubDir, "repo-settings.yaml")
	if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	tests := []struct {
		name       string
		configPath string
		want       string
	}{
		{
			name:       "directory path",
			configPath: githubDir,
			want:       filepath.Join(githubDir, ".env"),
		},
		{
			name:       "file in .github",
			configPath: configFile,
			want:       filepath.Join(githubDir, ".env"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveDotEnvPath(tt.configPath)
			if got != tt.want {
				t.Errorf("resolveDotEnvPath(%q) = %q, want %q", tt.configPath, got, tt.want)
			}
		})
	}
}

func TestWriteToEnvFile(t *testing.T) {
	tests := []struct {
		name            string
		existingContent string
		newValues       map[string]string
		wantContains    []string
		wantNotContain  []string
	}{
		{
			name:            "write to new file",
			existingContent: "",
			newValues: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			wantContains: []string{"KEY1=value1", "KEY2=value2", "# Added by provider:"},
		},
		{
			name:            "append to existing file",
			existingContent: "EXISTING=existing_value\n",
			newValues: map[string]string{
				"NEW_KEY": "new_value",
			},
			wantContains: []string{"EXISTING=existing_value", "NEW_KEY=new_value"},
		},
		{
			name:            "skip existing keys",
			existingContent: "KEY1=original\n",
			newValues: map[string]string{
				"KEY1": "new_value",
				"KEY2": "value2",
			},
			wantContains:   []string{"KEY1=original", "KEY2=value2"},
			wantNotContain: []string{"KEY1=new_value"},
		},
		{
			name:            "quote values with spaces",
			existingContent: "",
			newValues: map[string]string{
				"KEY1": "value with spaces",
			},
			wantContains: []string{`KEY1="value with spaces"`},
		},
		{
			name:            "escape quotes in values",
			existingContent: "",
			newValues: map[string]string{
				"KEY1": `value with "quotes"`,
			},
			wantContains: []string{`KEY1="value with \"quotes\""`},
		},
		{
			name:            "empty values map",
			existingContent: "EXISTING=value\n",
			newValues:       map[string]string{},
			wantContains:    []string{"EXISTING=value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "writeenv-test")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			envPath := filepath.Join(tmpDir, ".env")

			// Write existing content if any
			if tt.existingContent != "" {
				if err := os.WriteFile(envPath, []byte(tt.existingContent), 0644); err != nil {
					t.Fatalf("failed to write existing content: %v", err)
				}
			}

			// Call writeToEnvFile
			err = writeToEnvFile(envPath, tt.newValues)
			if err != nil {
				t.Fatalf("writeToEnvFile() error = %v", err)
			}

			// Read result
			content, err := os.ReadFile(envPath)
			if err != nil {
				// If no new values, file might not exist
				if len(tt.newValues) == 0 && tt.existingContent == "" {
					return
				}
				t.Fatalf("failed to read result: %v", err)
			}

			contentStr := string(content)

			// Check expected content
			for _, want := range tt.wantContains {
				if !strings.Contains(contentStr, want) {
					t.Errorf("writeToEnvFile() result missing %q\nGot:\n%s", want, contentStr)
				}
			}

			// Check content that should not be present
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(contentStr, notWant) {
					t.Errorf("writeToEnvFile() result should not contain %q\nGot:\n%s", notWant, contentStr)
				}
			}
		})
	}
}

func TestLoadFromProvider_NilConfig(t *testing.T) {
	result, err := LoadFromProvider(nil, nil, nil, "/tmp")
	if err != nil {
		t.Fatalf("LoadFromProvider() with nil config should not error, got: %v", err)
	}
	if result == nil {
		t.Fatal("LoadFromProvider() returned nil result")
	}
	if result.Values == nil {
		t.Fatal("LoadFromProvider() returned nil Values map")
	}
	if len(result.Values) != 0 {
		t.Errorf("LoadFromProvider() with nil config should return empty values, got %d", len(result.Values))
	}
}

func TestProviderResult(t *testing.T) {
	result := &ProviderResult{
		Values: map[string]string{
			"KEY1": "value1",
		},
		WrittenFile: true,
	}

	if result.Values["KEY1"] != "value1" {
		t.Errorf("ProviderResult.Values[KEY1] = %q, want %q", result.Values["KEY1"], "value1")
	}
	if !result.WrittenFile {
		t.Error("ProviderResult.WrittenFile = false, want true")
	}
}

func TestLoadFromProvider_UnknownProvider(t *testing.T) {
	cfg := &ProviderConfig{
		Name:   "unknown",
		Secret: "test",
	}
	_, err := LoadFromProvider(context.Background(), cfg, nil, "/tmp")
	if err == nil {
		t.Fatal("LoadFromProvider() with unknown provider should error")
	}
	if !strings.Contains(err.Error(), "unknown provider") {
		t.Errorf("LoadFromProvider() error = %q, want error containing 'unknown provider'", err.Error())
	}
}
