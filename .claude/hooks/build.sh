#!/bin/bash
set -euo pipefail

input=$(cat /dev/stdin)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

[[ "$file_path" == *.go ]] || exit 0

cd "$(dirname "$0")/../.."

if go build -o /tmp/http-server-build-check app/*.go 2>&1; then
  echo '{"systemMessage": "[Build] OK"}'
else
  echo '{"systemMessage": "[Build] FAILED"}'
  exit 1
fi
