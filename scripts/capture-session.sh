#!/bin/bash

SESSION_ID=$(jq -r '.session_id // empty')

[ -z "$SESSION_ID" ] && exit 0
[ -z "$CLAUDE_ENV_FILE" ] && exit 0

CURRENT_SESSION=""
if [ -f "$CLAUDE_ENV_FILE" ]; then
    CURRENT_SESSION=$(grep "^export CLAUDE_SESSION_ID=" "$CLAUDE_ENV_FILE" 2>/dev/null | tail -1 | cut -d"'" -f2)
fi

if [ -f "$CLAUDE_ENV_FILE" ]; then
    grep -v "^export CLAUDE_SESSION_ID=\|^export CLAUDE_PREV_SESSION_ID=" "$CLAUDE_ENV_FILE" > "$CLAUDE_ENV_FILE.tmp" 2>/dev/null || true
    mv "$CLAUDE_ENV_FILE.tmp" "$CLAUDE_ENV_FILE"
fi

if [ -n "$CURRENT_SESSION" ] && [ "$CURRENT_SESSION" != "$SESSION_ID" ]; then
    echo "export CLAUDE_PREV_SESSION_ID='$CURRENT_SESSION'" >> "$CLAUDE_ENV_FILE"
fi

echo "export CLAUDE_SESSION_ID='$SESSION_ID'" >> "$CLAUDE_ENV_FILE"
