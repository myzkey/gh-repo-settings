#!/bin/bash
# Generate Go types from OpenAPI schema using oapi-codegen
#
# This script generates Go types from the extracted OpenAPI subset schema.
# It only generates types (not client code), which can be used with the
# existing gh CLI wrapper.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OPENAPI_DIR="${PROJECT_ROOT}/internal/githubopenapi"
CONFIG_FILE="${OPENAPI_DIR}/oapi-codegen.yaml"
INPUT_FILE="${OPENAPI_DIR}/openapi-subset.json"
OUTPUT_FILE="${OPENAPI_DIR}/types.gen.go"

# oapi-codegen version
OAPI_CODEGEN_VERSION="v2.4.1"

# Check if input file exists
if [[ ! -f "$INPUT_FILE" ]]; then
    echo "Error: OpenAPI subset not found: $INPUT_FILE"
    echo "Run 'make extract-openapi' first to generate the subset."
    exit 1
fi

# Check if config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "Error: oapi-codegen config not found: $CONFIG_FILE"
    exit 1
fi

echo "Generating Go types from OpenAPI schema..."
echo "Input: $INPUT_FILE"
echo "Output: $OUTPUT_FILE"

# Run oapi-codegen from project root so relative paths in config work correctly
cd "$PROJECT_ROOT"
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION} --config "$CONFIG_FILE" "$INPUT_FILE"

echo "Done! Generated: $OUTPUT_FILE"

# Show summary
echo ""
echo "Summary:"
wc -l "$OUTPUT_FILE" | awk '{print "  Lines: " $1}'
grep -c "^type " "$OUTPUT_FILE" | xargs -I{} echo "  Types: {}" || true
