# Claude Coding

A Claude Code plugin that exports conversation threads to shareable links via GitHub Gists.

This is inspired by [ampcode.com](https://ampcode.com) threads.

## Features

- Exports Claude Code sessions to self-contained HTML files
- Creates shareable preview links via GitHub Gists
- **Session Linking**: Navigate between related sessions with Previous/Next links
  - When you use `/clear` to start a new session, it automatically links to the previous session
  - Exported gists include navigation to browse your session history within a single claude code terminal session.

## Installation

### Prerequisites
- [Go](https://go.dev/dl/) 1.20+
- Claude Code installed and set up
- `gh` CLI installed and authenticated (for Gist creation)

### Install as Claude Code Plugin
```bash
# 1. Install the CLI binary
go install github.com/priyanshujain/claude-coding/cmd/claude-coding@latest

# 2. Start Claude Code
claude

# 3. Add the marketplace from GitHub
/plugin marketplace add priyanshujain/claude-coding

# 4. Install the plugin
/plugin install claude-coding@priyanshujain
```

NOTE: Enable auto-updates for the plugin when adding marketplace and installing plugin to get the latest features and fixes.

## Usage

### Slash Command (Recommended)

After installing the plugin, use the `/share` command in Claude Code:

```
/claude-coding:share
```

This creates a GitHub Gist and returns a shareable preview link.

### Session Linking with /clear

When you use `/clear` to start a new conversation within the same project, the plugin automatically tracks session relationships:

1. Your sessions form a linked chain (Session A → Session B → Session C)
2. When you `/share`, the exported HTML includes navigation links
3. Viewers can browse through your session history using Previous/Next links
4. All linked session gists are automatically updated with the correct navigation


## How It Works

Claude Code stores conversation data in JSONL files at:
```
~/.claude/projects/{encoded-project-path}/{session-id}.jsonl
```

When you run `/share`:
1. The plugin finds your current session
2. Parses the conversation messages
3. Converts to a self-contained HTML file with syntax highlighting
4. Creates/updates a GitHub Gist
5. Returns a preview URL via [gistpreview.github.io](https://gistpreview.github.io)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, architecture details, and coding guidelines.


