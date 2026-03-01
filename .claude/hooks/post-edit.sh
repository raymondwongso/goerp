#!/usr/bin/env bash
# PostToolUse hook: fires after Write, Edit, or NotebookEdit on Go files.
# Runs gofmt and go vet on the modified file and injects findings back into
# the conversation via stderr so the senior-engineer can act on them.

set -euo pipefail

INPUT=$(cat)

FILE_PATH=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('file_path',''))" 2>/dev/null || true)

# Only act on .go files
if [[ -z "$FILE_PATH" || "$FILE_PATH" != *.go ]]; then
  exit 0
fi

# Resolve to absolute path if needed
if [[ ! "$FILE_PATH" = /* ]]; then
  FILE_PATH="$(pwd)/$FILE_PATH"
fi

if [[ ! -f "$FILE_PATH" ]]; then
  exit 0
fi

ISSUES=""

# --- gofmt check ---
GOFMT_DIFF=$(gofmt -l "$FILE_PATH" 2>/dev/null || true)
if [[ -n "$GOFMT_DIFF" ]]; then
  gofmt -w "$FILE_PATH" 2>/dev/null || true
  ISSUES="${ISSUES}[gofmt] Applied formatting to $FILE_PATH\n"
fi

# --- go vet check (on the package containing the file) ---
PKG_DIR=$(dirname "$FILE_PATH")
VET_OUTPUT=$(cd "$PKG_DIR" && go vet . 2>&1 || true)
if [[ -n "$VET_OUTPUT" ]]; then
  ISSUES="${ISSUES}[go vet] $VET_OUTPUT\n"
fi

if [[ -n "$ISSUES" ]]; then
  echo -e "⚠ Post-edit checks for $FILE_PATH:\n$ISSUES" >&2
fi

exit 0
