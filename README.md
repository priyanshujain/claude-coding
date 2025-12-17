# Claude Share

A Claude Code plugin that exports conversation threads to shareable HTML files.

## Features

- Exports current Claude Code session to a self-contained HTML file
- Clean, minimal UI inspired by [ampcode.com](https://ampcode.com) thread sharing
- Proper markdown rendering (headers, bold, code blocks, lists, links)
- Collapsible sections for thinking blocks and tool results
- Smart tool display:
  - **WebFetch/WebSearch**: Shows clickable URL + prompt
  - **Read/Glob/Grep**: Shows file path
  - **Bash**: Shows command description or truncated command
- Tool results grouped with their tool calls
- Auto-extracts title from first user message
- Detects system username and generates avatar initials

## Installation

### Install as Claude Code Plugin (Recommended)

```bash
# 1. Install the CLI binary
go install github.com/priyanshujain/claude-coding/cmd/claude-share@latest

# 2. Add the marketplace from GitHub
/plugin marketplace add priyanshujain/claude-coding

# 3. Install the plugin
/plugin install claude-share@priyanshujain
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/priyanshujain/claude-coding.git
cd claude-coding

# Build and install
make build
sudo make install

# Then install the plugin (inside Claude Code)
/plugin marketplace add priyanshujain/claude-coding
/plugin install claude-share@priyanshujain
```

## Usage

### CLI

```bash
# Export current directory's most recent session
claude-share

# Export specific project
claude-share --project /path/to/project

# Export specific session by ID
claude-share --session "abc123-session-id"

# Custom output path
claude-share --output my-thread.html

# Custom title and username
claude-share --title "My Thread" --username "John Doe"
```

### Slash Command

After installing the plugin, use the `/share` command in Claude Code:

```
/share
```

This will generate an HTML file in your current directory.

## How It Works

### Data Source

Claude Code stores conversation data in JSONL files at:
```
~/.claude/projects/{encoded-project-path}/{session-id}.jsonl
```

The project path is encoded by replacing `/` and `.` with `-`. For example:
- `/Users/pj/my-project` → `-Users-pj-my-project`

### JSONL Format

Each line in the JSONL file is a JSON object with:

```json
{
  "type": "user|assistant",
  "uuid": "message-uuid",
  "parentUuid": "parent-message-uuid",
  "sessionId": "session-id",
  "timestamp": "ISO-8601 timestamp",
  "message": {
    "role": "user|assistant",
    "content": "string or array of content blocks"
  }
}
```

Content blocks can be:
- `{"type": "text", "text": "..."}` - Plain text
- `{"type": "thinking", "thinking": "..."}` - Claude's thinking
- `{"type": "tool_use", "name": "...", "input": {...}}` - Tool invocation
- `{"type": "tool_result", "content": "..."}` - Tool output

### HTML Generation

The tool:
1. Finds the session file (by ID if provided, or most recent)
2. Parses messages and filters out `file-history-snapshot` entries
3. Merges tool results with their preceding tool calls
4. Converts markdown to HTML (headers, bold, code, lists, links)
5. Generates a self-contained HTML file with inline CSS

### Multi-Session Support

When multiple Claude Code sessions are running in parallel, the plugin uses a `SessionStart` hook to track which session invoked the `/share` command:

1. When a session starts, the hook captures the `session_id` from Claude Code
2. The session ID is persisted to `CLAUDE_ENV_FILE` (session-specific)
3. When `/share` runs, it passes `$CLAUDE_SESSION_ID` to the CLI
4. The CLI exports the correct session, not just the most recently modified one

This ensures reliable exports even with concurrent sessions in the same project.

## Project Structure

```
claude-coding/
├── .claude-plugin/
│   ├── plugin.json          # Plugin manifest
│   ├── marketplace.json     # Plugin marketplace for installation
│   └── scripts/
│       └── capture-session.sh  # SessionStart hook to capture session ID
├── commands/
│   └── share.md             # /share slash command
├── cmd/
│   └── claude-share/
│       └── main.go          # CLI entry point
├── internal/
│   ├── parser/
│   │   └── jsonl.go         # JSONL parsing & session discovery
│   ├── converter/
│   │   └── html.go          # Markdown → HTML conversion
│   └── template/
│       └── template.go      # HTML template
├── go.mod
├── Makefile
└── README.md
```

## HTML Output

The generated HTML includes:

- **Header**: Title (from first message) + username with avatar
- **Messages**:
  - User messages with pink avatar (initials)
  - Assistant messages with Claude icon
- **Collapsible sections**:
  - Thinking blocks (click to expand)
  - Tool results (click to expand)
- **Tool pills**: Compact display with icons for different tool types
- **Code blocks**: Syntax-highlighted with dark theme
- **Responsive design**: Max-width 720px, centered

## Development

```bash
# Build
make build

# Install locally
make install

# Clean build artifacts
make clean

# Test
./bin/claude-share --project "$PWD" --output test.html
open test.html
```

## License

Apache 2.0
