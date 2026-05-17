#!/bin/bash

input=$(cat /dev/stdin)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')
[[ "$file_path" == *.go ]] || exit 0

LOG=/tmp/http-server/unit-test.log
cd "$(dirname "$0")/../.." && go test -v ./app/... > "$LOG" 2>&1
UNIT_EXIT=$?
if [[ $UNIT_EXIT -ne 0 ]]; then
  echo "{\"systemMessage\": \"[Test] Unit tests failed — details in $LOG\"}"
  exit 1
fi

echo "{\"systemMessage\": \"[Test] All unit tests passed\"}"


