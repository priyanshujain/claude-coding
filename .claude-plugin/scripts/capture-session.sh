#!/bin/bash
# Capture session ID from Claude Code hook input and persist to CLAUDE_ENV_FILE
# This script runs during SessionStart and makes CLAUDE_SESSION_ID available
# to all subsequent bash commands in this session.

INPUT=$(cat)

SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')
ENV_FILE=$(echo "$INPUT" | jq -r '.CLAUDE_ENV_FILE // empty')

if [ -n "$SESSION_ID" ] && [ -n "$ENV_FILE" ]; then
    echo "export CLAUDE_SESSION_ID='$SESSION_ID'" >> "$ENV_FILE"
fi

exit 0
