---
description: Export current thread to shareable HTML (Gist)
---

Export the current conversation thread to a GitHub Gist for sharing.

Run the claude-share CLI tool with --gist flag:
!claude-share --project "$PWD" --session "$CLAUDE_SESSION_ID" --gist

After generating, tell the user the gistpreview.github.io URL that was output.
