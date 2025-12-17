package converter

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/priyanshujain/claude-coding/internal/parser"
	"github.com/priyanshujain/claude-coding/internal/template"
)

type Config struct {
	Title        string
	Username     string
	UserInitials string
	ProjectPath  string
}

var currentProjectPath string

func Convert(messages []parser.Message, cfg Config) string {
	currentProjectPath = cfg.ProjectPath
	messages = mergeToolResults(messages)

	var messagesHTML strings.Builder
	for _, msg := range messages {
		messagesHTML.WriteString(renderMessage(msg, cfg))
	}

	result := template.HTMLTemplate
	result = strings.ReplaceAll(result, "TITLE_PLACEHOLDER", html.EscapeString(cfg.Title))
	result = strings.ReplaceAll(result, "USERNAME_PLACEHOLDER", html.EscapeString(cfg.Username))
	result = strings.ReplaceAll(result, "INITIALS_PLACEHOLDER", html.EscapeString(cfg.UserInitials))
	result = strings.ReplaceAll(result, "MESSAGES_PLACEHOLDER", messagesHTML.String())

	return result
}

func mergeToolResults(messages []parser.Message) []parser.Message {
	// First pass: build maps of tool_use IDs to tool names and inputs
	toolNames := make(map[string]string)
	toolInputs := make(map[string]string)
	for _, msg := range messages {
		for _, block := range msg.Blocks {
			if block.Type == "tool_use" && block.ToolUseID != "" {
				toolNames[block.ToolUseID] = block.ToolName
				toolInputs[block.ToolUseID] = block.ToolInput
			}
		}
	}

	// Second pass: merge tool_results into assistant messages
	// Insert each result right after its corresponding tool_use
	var result []parser.Message
	for i := 0; i < len(messages); i++ {
		msg := messages[i]

		if msg.Role == "user" && len(msg.Blocks) > 0 && msg.Blocks[0].Type == "tool_result" {
			if len(result) > 0 && result[len(result)-1].Role == "assistant" {
				lastAssistant := &result[len(result)-1]

				// For each tool_result, insert it after its matching tool_use
				for _, resultBlock := range msg.Blocks {
					if resultBlock.ToolUseID != "" {
						resultBlock.ToolName = toolNames[resultBlock.ToolUseID]
						resultBlock.ToolInput = toolInputs[resultBlock.ToolUseID]
					}

					// Find the matching tool_use and insert result after it
					inserted := false
					var newBlocks []parser.ContentBlock
					for _, block := range lastAssistant.Blocks {
						newBlocks = append(newBlocks, block)
						if block.Type == "tool_use" && block.ToolUseID == resultBlock.ToolUseID {
							newBlocks = append(newBlocks, resultBlock)
							inserted = true
						}
					}
					if inserted {
						lastAssistant.Blocks = newBlocks
					} else {
						// Fallback: append at end if no match found
						lastAssistant.Blocks = append(lastAssistant.Blocks, resultBlock)
					}
				}
				continue
			}
		}
		result = append(result, msg)
	}
	return result
}

func renderMessage(msg parser.Message, cfg Config) string {
	var content strings.Builder

	for _, block := range msg.Blocks {
		content.WriteString(renderBlock(block))
	}

	// Skip empty messages (e.g., when all content was filtered out)
	if strings.TrimSpace(content.String()) == "" {
		return ""
	}

	if msg.Role == "user" {
		return `<div class="message user">
<span class="avatar">` + html.EscapeString(cfg.UserInitials) + `</span>
<div class="message-content">` + content.String() + `</div>
</div>`
	}

	return `<div class="message assistant">
<span class="avatar">` + template.ClaudeIcon + `</span>
<div class="message-content">` + content.String() + `</div>
</div>`
}

func renderBlock(block parser.ContentBlock) string {
	switch block.Type {
	case "text":
		content := strings.TrimSpace(block.Content)
		if content == "" {
			return ""
		}
		// Skip system-injected interrupt message
		if content == "[Request interrupted by user for tool use]" {
			return ""
		}
		return `<div class="text-block">` + formatText(block.Content) + `</div>`

	case "thinking":
		return `<div class="collapsible">
<div class="collapsible-header"><span class="chevron">▶</span> Thinking</div>
<div class="collapsible-content">` + html.EscapeString(block.Content) + `</div>
</div>`

	case "tool_use":
		return renderToolUse(block.Content)

	case "tool_result":
		return renderToolResult(block)
	}

	return ""
}

func renderToolUse(content string) string {
	parts := strings.SplitN(content, "\n", 2)
	toolName := parts[0]
	toolInput := ""
	if len(parts) > 1 {
		toolInput = parts[1]
	}

	icon := getToolIcon(toolName)

	if strings.Contains(toolName, "WebFetch") || strings.Contains(toolName, "WebSearch") {
		return renderWebTool(toolName, toolInput, icon)
	}

	if strings.Contains(toolName, "Read") {
		return renderReadTool(toolName, toolInput, icon)
	}

	if strings.Contains(toolName, "Bash") {
		return renderBashTool(toolName, toolInput, icon)
	}

	if strings.Contains(toolName, "Edit") {
		return renderEditTool(toolName, toolInput, icon)
	}

	if strings.Contains(toolName, "Glob") {
		return renderGlobTool(toolName, toolInput)
	}

	return `<div class="tool-block">
<div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div>
</div>`
}

func renderWebTool(toolName, input, icon string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	url, _ := data["url"].(string)
	prompt, _ := data["prompt"].(string)
	query, _ := data["query"].(string)

	var info strings.Builder
	if url != "" {
		info.WriteString(`<div><a href="` + html.EscapeString(url) + `" target="_blank">` + html.EscapeString(url) + `</a></div>`)
	}
	if query != "" {
		info.WriteString(`<div style="margin-top:4px;color:#888;">` + html.EscapeString(query) + `</div>`)
	}
	if prompt != "" && len(prompt) < 200 {
		info.WriteString(`<div style="margin-top:4px;color:#888;">` + html.EscapeString(prompt) + `</div>`)
	}

	return `<div class="tool-block">
<div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div>
<div class="tool-info">` + info.String() + `</div>
</div>`
}

func renderReadTool(toolName, input, icon string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	filePath, _ := data["file_path"].(string)
	if filePath == "" {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	displayPath := getRelativePath(filePath)

	return `<div class="tool-block">
<div class="tool-pill" title="` + html.EscapeString(filePath) + `">` + icon + ` ` + html.EscapeString(displayPath) + `</div>
</div>`
}

func renderBashTool(toolName, input, icon string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	cmd, _ := data["command"].(string)
	desc, _ := data["description"].(string)

	// Show description in pill, command in code block
	pillText := desc
	if pillText == "" {
		pillText = "Bash"
	}

	var result strings.Builder
	result.WriteString(`<div class="tool-block">`)
	result.WriteString(`<div class="tool-pill">` + icon + ` ` + html.EscapeString(pillText) + `</div>`)

	if cmd != "" {
		result.WriteString(`<div class="bash-command"><code>` + html.EscapeString(cmd) + `</code></div>`)
	}

	result.WriteString(`</div>`)
	return result.String()
}

func renderEditTool(toolName, input, icon string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	filePath, _ := data["file_path"].(string)
	oldString, _ := data["old_string"].(string)
	newString, _ := data["new_string"].(string)

	if filePath == "" {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	// Get relative path for display
	displayPath := getRelativePath(filePath)

	var result strings.Builder
	result.WriteString(`<div class="tool-block">`)
	result.WriteString(`<div class="tool-pill" title="` + html.EscapeString(filePath) + `">` + icon + ` ` + html.EscapeString(displayPath) + `</div>`)

	// Show diff if we have old and new strings
	if oldString != "" || newString != "" {
		result.WriteString(`<div class="diff-block">`)

		// Show removed lines (red)
		if oldString != "" {
			oldLines := strings.Split(oldString, "\n")
			for _, line := range oldLines {
				result.WriteString(`<div class="diff-line diff-removed">- ` + html.EscapeString(line) + `</div>`)
			}
		}

		// Show added lines (green)
		if newString != "" {
			newLines := strings.Split(newString, "\n")
			for _, line := range newLines {
				result.WriteString(`<div class="diff-line diff-added">+ ` + html.EscapeString(line) + `</div>`)
			}
		}

		result.WriteString(`</div>`)
	}

	result.WriteString(`</div>`)
	return result.String()
}

func renderGlobTool(toolName, input string) string {
	searchIcon := `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>`

	var data map[string]any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return `<div class="tool-block"><div class="tool-pill">` + searchIcon + ` Search</div></div>`
	}

	pattern, _ := data["pattern"].(string)
	if pattern == "" {
		return `<div class="tool-block"><div class="tool-pill">` + searchIcon + ` Search</div></div>`
	}

	return `<div class="tool-block">
<div class="tool-pill">` + searchIcon + ` ` + html.EscapeString(pattern) + `</div>
</div>`
}

func renderToolResult(block parser.ContentBlock) string {
	toolName := block.ToolName
	content := block.Content

	// Handle error/rejected tool calls
	if block.IsError {
		return `<div class="tool-result-error">` + html.EscapeString(content) + `</div>`
	}

	// Skip Edit results - the diff in tool_use already shows the changes
	if toolName == "Edit" {
		return ""
	}

	// Handle Glob tool results specially - show file paths clearly
	if toolName == "Glob" {
		return renderGlobResult(content)
	}

	// Handle Grep tool results specially
	if toolName == "Grep" {
		return renderGrepResult(content)
	}

	// Handle Read tool results - strip line numbers and add syntax highlighting
	if toolName == "Read" {
		content = stripLineNumbers(content)
		lang := getLanguageFromInput(block.ToolInput)
		truncated := content
		if len(truncated) > 2000 {
			truncated = truncated[:2000] + "\n... (truncated)"
		}
		return `<div class="collapsible tool-result">
<div class="collapsible-header"><span class="chevron">▶</span> Read Result</div>
<div class="collapsible-content"><pre><code class="language-` + lang + `">` + html.EscapeString(truncated) + `</code></pre></div>
</div>`
	}

	// Default: show as collapsible with tool name
	truncated := content
	if len(truncated) > 2000 {
		truncated = truncated[:2000] + "\n... (truncated)"
	}

	headerText := "Result"
	if toolName != "" {
		headerText = toolName + " Result"
	}

	return `<div class="collapsible tool-result">
<div class="collapsible-header"><span class="chevron">▶</span> ` + html.EscapeString(headerText) + `</div>
<div class="collapsible-content"><pre>` + html.EscapeString(truncated) + `</pre></div>
</div>`
}

func renderGlobResult(content string) string {
	if strings.TrimSpace(content) == "" {
		return `<div class="search-result"><span class="search-result-count">No files found</span></div>`
	}

	// Split content into file paths
	paths := strings.Split(strings.TrimSpace(content), "\n")
	var validPaths []string
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path != "" {
			validPaths = append(validPaths, path)
		}
	}

	if len(validPaths) == 0 {
		return `<div class="search-result"><span class="search-result-count">No files found</span></div>`
	}

	// Build file list
	var result strings.Builder
	result.WriteString(`<div class="search-result">`)
	result.WriteString(fmt.Sprintf(`<span class="search-result-count">Found %d files</span>`, len(validPaths)))
	result.WriteString(`<div class="search-result-list">`)

	for _, path := range validPaths {
		displayPath := getRelativePath(path)
		result.WriteString(`<div class="search-result-item" title="` + html.EscapeString(path) + `">` + html.EscapeString(displayPath) + `</div>`)
	}

	result.WriteString(`</div></div>`)
	return result.String()
}

func renderGrepResult(content string) string {
	if strings.TrimSpace(content) == "" {
		return `<div class="tool-result-inline">No matches found</div>`
	}

	// Split content into file paths (assuming files_with_matches mode)
	paths := strings.Split(strings.TrimSpace(content), "\n")
	if len(paths) == 0 {
		return `<div class="tool-result-inline">No matches found</div>`
	}

	// Check if it looks like file paths (simple heuristic)
	looksLikeFilePaths := true
	for _, path := range paths {
		if strings.Contains(path, ":") && !strings.HasPrefix(path, "/") {
			looksLikeFilePaths = false
			break
		}
	}

	if looksLikeFilePaths && len(paths) <= 20 {
		// Show as file list similar to Glob
		var result strings.Builder
		result.WriteString(`<div class="tool-result-files">`)

		for _, path := range paths {
			path = strings.TrimSpace(path)
			if path == "" {
				continue
			}
			displayPath := getRelativePath(path)
			result.WriteString(`<div class="file-path" title="` + html.EscapeString(path) + `">` + html.EscapeString(displayPath) + `</div>`)
		}

		result.WriteString(`</div>`)
		return result.String()
	}

	// Otherwise show as truncated code block
	truncated := content
	if len(truncated) > 2000 {
		truncated = truncated[:2000] + "\n... (truncated)"
	}

	return `<div class="collapsible tool-result">
<div class="collapsible-header"><span class="chevron">▶</span> Grep Result</div>
<div class="collapsible-content"><pre>` + html.EscapeString(truncated) + `</pre></div>
</div>`
}

func stripLineNumbers(content string) string {
	// Strip line number prefixes like "     1→\t" from Read tool output
	lines := strings.Split(content, "\n")
	var result []string
	lineNumPattern := regexp.MustCompile(`^\s*\d+→\t?`)
	for _, line := range lines {
		result = append(result, lineNumPattern.ReplaceAllString(line, ""))
	}
	return strings.Join(result, "\n")
}

func getLanguageFromInput(toolInput string) string {
	// Extract file_path from tool input JSON and determine language
	var data map[string]any
	if err := json.Unmarshal([]byte(toolInput), &data); err != nil {
		return "plaintext"
	}
	filePath, _ := data["file_path"].(string)
	if filePath == "" {
		return "plaintext"
	}

	// Map file extensions to Prism language names
	ext := strings.ToLower(filePath)
	if idx := strings.LastIndex(ext, "."); idx >= 0 {
		ext = ext[idx:]
	}

	langMap := map[string]string{
		".go":    "go",
		".py":    "python",
		".js":    "javascript",
		".ts":    "typescript",
		".tsx":   "typescript",
		".jsx":   "javascript",
		".json":  "json",
		".yaml":  "yaml",
		".yml":   "yaml",
		".md":    "markdown",
		".sh":    "bash",
		".bash":  "bash",
		".zsh":   "bash",
		".rs":    "rust",
		".html":  "markup",
		".xml":   "markup",
		".css":   "css",
		".sql":   "sql",
		".rb":    "ruby",
		".java":  "java",
		".c":     "c",
		".cpp":   "cpp",
		".h":     "c",
		".hpp":   "cpp",
	}

	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "plaintext"
}

func getRelativePath(fullPath string) string {
	// Use project path to get relative path
	if currentProjectPath != "" && strings.HasPrefix(fullPath, currentProjectPath) {
		rel := strings.TrimPrefix(fullPath, currentProjectPath)
		return strings.TrimPrefix(rel, "/")
	}
	// Fallback: return last 3 path components
	parts := strings.Split(fullPath, "/")
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return fullPath
}

func getToolIcon(toolName string) string {
	switch {
	case strings.Contains(toolName, "WebSearch"), strings.Contains(toolName, "WebFetch"):
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>`
	case strings.Contains(toolName, "Read"), strings.Contains(toolName, "Glob"), strings.Contains(toolName, "Grep"):
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>`
	case strings.Contains(toolName, "Write"), strings.Contains(toolName, "Edit"):
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>`
	case strings.Contains(toolName, "Bash"):
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>`
	case strings.Contains(toolName, "Task"), strings.Contains(toolName, "Todo"):
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><rect x="8" y="2" width="8" height="4" rx="1" ry="1"/></svg>`
	default:
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>`
	}
}

var (
	codeBlockRegex  = regexp.MustCompile("(?s)```(\\w*)\\n?(.*?)```")
	inlineCodeRegex = regexp.MustCompile("`([^`\n]+)`")
	boldRegex       = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	linkRegex       = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	h3Regex         = regexp.MustCompile(`(?m)^### (.+)$`)
	h2Regex         = regexp.MustCompile(`(?m)^## (.+)$`)
	h1Regex         = regexp.MustCompile(`(?m)^# (.+)$`)
)

func formatText(text string) string {
	var codeBlocks []string
	text = codeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := codeBlockRegex.FindStringSubmatch(match)
		code := html.EscapeString(parts[2])
		placeholder := fmt.Sprintf("__CODEBLOCK_%d__", len(codeBlocks))
		codeBlocks = append(codeBlocks, "<pre><code>"+code+"</code></pre>")
		return placeholder
	})

	var inlineCode []string
	text = inlineCodeRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := inlineCodeRegex.FindStringSubmatch(match)
		placeholder := fmt.Sprintf("__INLINECODE_%d__", len(inlineCode))
		inlineCode = append(inlineCode, "<code>"+html.EscapeString(parts[1])+"</code>")
		return placeholder
	})

	text = html.EscapeString(text)

	for i, code := range codeBlocks {
		text = strings.Replace(text, fmt.Sprintf("__CODEBLOCK_%d__", i), code, 1)
	}
	for i, code := range inlineCode {
		text = strings.Replace(text, fmt.Sprintf("__INLINECODE_%d__", i), code, 1)
	}

	text = boldRegex.ReplaceAllString(text, "<strong>$1</strong>")
	text = linkRegex.ReplaceAllString(text, `<a href="$2" target="_blank">$1</a>`)
	text = h3Regex.ReplaceAllString(text, "<h4>$1</h4>")
	text = h2Regex.ReplaceAllString(text, "<h3>$1</h3>")
	text = h1Regex.ReplaceAllString(text, "<h2>$1</h2>")

	var result strings.Builder
	lines := strings.Split(text, "\n")
	inList := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "- ") {
			if !inList {
				result.WriteString("<ul>")
				inList = true
			}
			result.WriteString("<li>" + strings.TrimPrefix(trimmed, "- ") + "</li>")
			continue
		}

		if inList {
			result.WriteString("</ul>")
			inList = false
		}

		result.WriteString(line + "\n")
	}

	if inList {
		result.WriteString("</ul>")
	}

	return strings.TrimSpace(result.String())
}
