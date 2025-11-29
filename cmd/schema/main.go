package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/myzkey/gh-repo-settings/internal/config"
)

func main() {
	r := new(jsonschema.Reflector)
	r.ExpandedStruct = true

	schema := r.Reflect(&config.Config{})
	schema.ID = "https://raw.githubusercontent.com/myzkey/gh-repo-settings/main/schema.json"
	schema.Title = "gh-repo-settings configuration"
	schema.Description = "Configuration schema for gh-repo-settings (gh rset) - GitHub repository settings management tool"

	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(out))
}
