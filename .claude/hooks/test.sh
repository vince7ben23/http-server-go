#!/bin/bash

BASE="http://localhost:4221"
PASS=0
FAIL=0

check() {
  local desc="$1" expected="$2"
  shift 2
  local actual
  actual=$(curl -s --max-time 3 -o /dev/null -w "%{http_code}" "$@")
  if [[ "$actual" == "$expected" ]]; then
    PASS=$((PASS + 1))
  else
    FAIL=$((FAIL + 1))
    echo "[FAIL] $desc — expected $expected, got $actual" >&2
  fi
}

input=$(cat /dev/stdin)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')
[[ "$file_path" == *.go ]] || exit 0

SERVER_STARTED=0
if ! curl -sf --max-time 2 "$BASE/" -o /dev/null; then
  /tmp/http-server-build-check --directory /tmp/ >/tmp/hook-server.log 2>&1 &
  SERVER_PID=$!
  SERVER_STARTED=1
  # Wait for server to be ready (up to 5s)
  for i in {1..10}; do
    sleep 0.5
    curl -sf --max-time 1 "$BASE/" -o /dev/null && break
  done
fi

check "GET /"                       200  "$BASE/"
check "GET /echo/hello"             200  "$BASE/echo/hello"
check "GET /user-agent"             200  -H "User-Agent: test-agent" "$BASE/user-agent"
check "GET /nonexistent → 404"      404  "$BASE/nonexistent"
check "POST /files/test.txt → 201"  201  -X POST -d "hello" "$BASE/files/test.txt"
check "GET /files/test.txt → 200"   200  "$BASE/files/test.txt"

if [[ $FAIL -eq 0 ]]; then
  echo "{\"systemMessage\": \"[Test] All $PASS passed\"}"
else
  echo "{\"systemMessage\": \"[Test] $FAIL failed, $PASS passed — check stderr\"}"
fi

if [[ $SERVER_STARTED -eq 1 ]]; then
  kill "$SERVER_PID" 2>/dev/null
  exit 0
fi


