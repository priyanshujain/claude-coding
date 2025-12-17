---
description: Export current thread to shareable HTML
---

Export the current conversation thread to an HTML file for sharing.

Run the claude-share CLI tool to generate the HTML:
!claude-share --project "$PWD" --session "$CLAUDE_SESSION_ID"

After generating, tell the user the output file path.
