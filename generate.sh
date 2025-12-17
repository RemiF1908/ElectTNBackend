#!/bin/bash

# Generate Go code from OpenAPI specification
echo "Generating Go code from openapi.yaml..."

# Create api directory if it doesn't exist
mkdir -p api

# Generate types, server interfaces, client, and spec with response type suffix
oapi-codegen --config oapi-codegen.yaml api/openapi.yaml

echo "Code generation completed successfully!"