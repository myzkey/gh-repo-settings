//go:build ignore

// Extract a subset of GitHub OpenAPI schema containing only the required endpoints.
//
// Usage:
//
//	go run scripts/extract-openapi-subset.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Required endpoints for this project
var requiredPaths = []string{
	// Repository
	"/repos/{owner}/{repo}",
	// Topics
	"/repos/{owner}/{repo}/topics",
	// Labels
	"/repos/{owner}/{repo}/labels",
	"/repos/{owner}/{repo}/labels/{name}",
	// Branch protection
	"/repos/{owner}/{repo}/branches/{branch}/protection",
	// Actions secrets
	"/repos/{owner}/{repo}/actions/secrets",
	"/repos/{owner}/{repo}/actions/secrets/public-key",
	"/repos/{owner}/{repo}/actions/secrets/{secret_name}",
	// Actions variables
	"/repos/{owner}/{repo}/actions/variables",
	"/repos/{owner}/{repo}/actions/variables/{name}",
	// Actions permissions
	"/repos/{owner}/{repo}/actions/permissions",
	"/repos/{owner}/{repo}/actions/permissions/selected-actions",
	"/repos/{owner}/{repo}/actions/permissions/workflow",
	// Pages
	"/repos/{owner}/{repo}/pages",
}

// Component types to track
var componentTypes = []string{"schemas", "parameters", "responses", "examples", "headers"}

// Schema renames to avoid conflicts (original name -> new name)
var schemaRenames = map[string]string{
	"page": "github-page",
}

func main() {
	inputFile := flag.String("input", "internal/githubopenapi/github-openapi-full.json", "Input OpenAPI schema file")
	outputFile := flag.String("output", "internal/githubopenapi/openapi-subset.json", "Output subset schema file")
	listPaths := flag.Bool("list-paths", false, "List required paths and exit")
	flag.Parse()

	if *listPaths {
		fmt.Println("Required paths:")
		for _, p := range requiredPaths {
			fmt.Printf("  %s\n", p)
		}
		return
	}

	// Read input file
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'make fetch-openapi' first to download the schema.")
		os.Exit(1)
	}

	fmt.Printf("Reading: %s\n", *inputFile)

	var spec map[string]interface{}
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Extracting %d paths...\n", len(requiredPaths))
	subset := extractSubset(spec, requiredPaths)

	schemas := getSchemas(subset)
	fmt.Printf("Collected %d schemas\n", len(schemas))

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(*outputFile), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Write output with indentation for readability
	out, err := json.MarshalIndent(subset, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFile, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Writing: %s\n", *outputFile)
	fmt.Println("Done!")
}

func extractSubset(spec map[string]interface{}, paths []string) map[string]interface{} {
	// Initialize subset structure
	info := getMap(spec, "info")
	subset := map[string]interface{}{
		"openapi": spec["openapi"],
		"info": map[string]interface{}{
			"title":       "GitHub REST API (Subset)",
			"description": "Subset of GitHub REST API for gh-repo-settings",
			"version":     info["version"],
		},
		"paths":      map[string]interface{}{},
		"components": map[string]interface{}{},
	}

	if servers, ok := spec["servers"]; ok {
		subset["servers"] = servers
	}

	// Initialize all component maps
	subsetComponents := subset["components"].(map[string]interface{})
	for _, compType := range componentTypes {
		subsetComponents[compType] = map[string]interface{}{}
	}

	specPaths := getMap(spec, "paths")
	subsetPaths := subset["paths"].(map[string]interface{})

	// Track all needed components: map[componentType]map[name]bool
	neededComponents := make(map[string]map[string]bool)
	for _, compType := range componentTypes {
		neededComponents[compType] = make(map[string]bool)
	}

	// Collect paths and find initial refs
	for _, path := range paths {
		pathItem, ok := specPaths[path]
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: Path '%s' not found in spec\n", path)
			continue
		}

		subsetPaths[path] = pathItem

		// Find all $refs in this path
		findRefs(pathItem, neededComponents)
	}

	// Recursively collect all components and their dependencies
	components := getMap(spec, "components")
	collectAllComponents(components, neededComponents, subsetComponents)

	// Rename conflicting schemas and update references
	renameSchemas(subset)

	// Sort schemas alphabetically
	if schemas, ok := subsetComponents["schemas"].(map[string]interface{}); ok {
		sortedSchemas := make(map[string]interface{})
		var schemaNames []string
		for name := range schemas {
			schemaNames = append(schemaNames, name)
		}
		sort.Strings(schemaNames)
		for _, name := range schemaNames {
			sortedSchemas[name] = schemas[name]
		}
		subsetComponents["schemas"] = sortedSchemas
	}

	return subset
}

// renameSchemas renames schemas that conflict with parameter names and updates all references
func renameSchemas(spec map[string]interface{}) {
	components := getMap(spec, "components")
	schemas := getMap(components, "schemas")

	// Rename schemas
	for oldName, newName := range schemaRenames {
		if schema, ok := schemas[oldName]; ok {
			// Add x-go-name to preserve Go type name
			if schemaMap, ok := schema.(map[string]interface{}); ok {
				schemaMap["x-go-name"] = "GitHubPage"
			}
			schemas[newName] = schema
			delete(schemas, oldName)
			fmt.Printf("Renamed schema '%s' to '%s'\n", oldName, newName)
		}
	}

	// Update all $ref references in the entire spec
	updateRefs(spec)
}

// updateRefs recursively updates $ref values for renamed schemas
func updateRefs(obj interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		if ref, ok := v["$ref"].(string); ok {
			for oldName, newName := range schemaRenames {
				oldRef := "#/components/schemas/" + oldName
				newRef := "#/components/schemas/" + newName
				if ref == oldRef {
					v["$ref"] = newRef
				}
			}
		}
		for _, val := range v {
			updateRefs(val)
		}
	case []interface{}:
		for _, item := range v {
			updateRefs(item)
		}
	}
}

// findRefs recursively finds all $ref values and adds them to neededComponents
func findRefs(obj interface{}, neededComponents map[string]map[string]bool) {
	switch v := obj.(type) {
	case map[string]interface{}:
		if ref, ok := v["$ref"].(string); ok {
			// Parse $ref like "#/components/schemas/repository"
			if len(ref) > 13 && ref[:13] == "#/components/" {
				rest := ref[13:]
				for i, c := range rest {
					if c == '/' {
						compType := rest[:i]
						name := rest[i+1:]
						if _, ok := neededComponents[compType]; ok {
							neededComponents[compType][name] = true
						}
						break
					}
				}
			}
		}
		for _, val := range v {
			findRefs(val, neededComponents)
		}
	case []interface{}:
		for _, item := range v {
			findRefs(item, neededComponents)
		}
	}
}

func collectAllComponents(specComponents map[string]interface{}, neededComponents map[string]map[string]bool, subsetComponents map[string]interface{}) {
	// Keep collecting until no new components are found
	for {
		foundNew := false

		for _, compType := range componentTypes {
			specCompMap := getMap(specComponents, compType)
			subsetCompMap := subsetComponents[compType].(map[string]interface{})
			needed := neededComponents[compType]

			for name := range needed {
				if _, exists := subsetCompMap[name]; exists {
					continue
				}

				comp, ok := specCompMap[name]
				if !ok {
					continue
				}

				subsetCompMap[name] = comp
				foundNew = true

				// Find refs in this component
				findRefs(comp, neededComponents)
			}
		}

		if !foundNew {
			break
		}
	}
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if vm, ok := v.(map[string]interface{}); ok {
			return vm
		}
	}
	return make(map[string]interface{})
}

func getSchemas(spec map[string]interface{}) map[string]interface{} {
	components := getMap(spec, "components")
	return getMap(components, "schemas")
}
