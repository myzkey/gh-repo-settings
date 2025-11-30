package cmd

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

// marshalYAML marshals data to YAML with 2-space indentation
func marshalYAML(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
