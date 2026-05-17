#!/bin/bash
set -euo pipefail

input=$(cat /dev/stdin)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

[[ "$file_path" == *.go ]] || exit 0

LOG=/tmp/http-server/build.log
cd "$(dirname "$0")/../.." && go build -o /tmp/http-server-build-check app/*.go > "$LOG" 2>&1
BUILD_EXIT=$?
if [[ $BUILD_EXIT -eq 0 ]]; then
  echo '{"systemMessage": "[Build] OK"}'
else
  echo "{\"systemMessage\": \"[Build] FAILED — check details in $LOG\"}"
  exit 1
fi
