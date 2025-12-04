#!/bin/bash
# Fetch GitHub REST API OpenAPI schema
#
# This script downloads the official GitHub REST API OpenAPI specification
# from the github/rest-api-description repository.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="${PROJECT_ROOT}/internal/githubopenapi"
OUTPUT_FILE="${OUTPUT_DIR}/github-openapi-full.json"

# GitHub OpenAPI schema URL (OpenAPI 3.0 JSON version for faster parsing)
# Note: descriptions-next/ contains OpenAPI 3.1, descriptions/ contains OpenAPI 3.0
OPENAPI_URL="https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json"

mkdir -p "$OUTPUT_DIR"

echo "Downloading GitHub OpenAPI schema..."
echo "URL: $OPENAPI_URL"

if command -v curl &> /dev/null; then
    curl -fsSL "$OPENAPI_URL" -o "$OUTPUT_FILE"
elif command -v wget &> /dev/null; then
    wget -q "$OPENAPI_URL" -O "$OUTPUT_FILE"
else
    echo "Error: curl or wget is required"
    exit 1
fi

echo "Downloaded to: $OUTPUT_FILE"
echo "File size: $(wc -l < "$OUTPUT_FILE") lines"
