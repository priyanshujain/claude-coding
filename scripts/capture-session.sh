#!/bin/bash
# Capture session ID from Claude Code hook input and persist to CLAUDE_ENV_FILE
# This script runs during SessionStart and makes CLAUDE_SESSION_ID available
# to all subsequent bash commands in this session.

INPUT=$(cat)

SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')

if [ -n "$SESSION_ID" ] && [ -n "$CLAUDE_ENV_FILE" ]; then
    if [ -f "$CLAUDE_ENV_FILE" ]; then
        grep -v "^export CLAUDE_SESSION_ID=" "$CLAUDE_ENV_FILE" > "$CLAUDE_ENV_FILE.tmp" 2>/dev/null || true
        mv "$CLAUDE_ENV_FILE.tmp" "$CLAUDE_ENV_FILE"
    fi
    echo "export CLAUDE_SESSION_ID='$SESSION_ID'" >> "$CLAUDE_ENV_FILE"
fi

exit 0
