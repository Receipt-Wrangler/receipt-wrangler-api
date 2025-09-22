#!/bin/bash

# Function to display usage
show_usage() {
    echo "Usage: $0 <platform> <output-dir>"
    echo ""
    echo "Arguments:"
    echo "  platform     Either 'desktop', 'mobile', or 'mcp'"
    echo "  output-dir   Directory path for generated code output"
    echo ""
    echo "Example:"
    echo "  $0 desktop /home/user/project/src/open-api"
    exit 1
}

# Check if correct number of arguments is provided
if [ $# -ne 2 ]; then
    echo "Error: Exactly 2 arguments are required"
    show_usage
fi

platform=$1
output_dir=$2

# Validate platform argument
if [ "$platform" != "desktop" ] && [ "$platform" != "mobile" ] && [ "$platform" != "mcp" ]; then
    echo "Error: Platform must be either 'desktop', 'mobile', or 'mcp'"
    show_usage
fi

# Set generator based on platform
if [ "$platform" = "desktop" ]; then
    generator="typescript-angular"
elif [ "$platform" = "mcp" ]; then
    generator="typescript"
else
    generator="dart-dio"
fi

# Execute the OpenAPI generator command
npx @openapitools/openapi-generator-cli generate \
    -i swagger.yml \
    -g "$generator" \
    -o "$output_dir"

# Check if command executed successfully
if [ $? -eq 0 ]; then
    echo "API code successfully generated in: $output_dir"
else
    echo "Error: API code generation failed"
    exit 1
fi
