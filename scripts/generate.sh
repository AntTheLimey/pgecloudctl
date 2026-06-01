#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "Installing oapi-codegen..."
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

echo "Generating API client..."
oapi-codegen -config openapi/generate.yaml openapi/pgedge.yaml

echo "Done. Generated files in internal/api/"
