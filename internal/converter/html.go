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
	Title          string
	Username       string
	UserInitials   string
	ProjectPath    string
	PrevSessionURL string
	NextSessionURL string
}

var currentProjectPath string

func Convert(messages []parser.Message, cfg Config) string {
	currentProjectPath = cfg.ProjectPath
	messages = mergeToolResults(messages)
	messages = mergeBashMessages(messages)

	var messagesHTML strings.Builder
	for _, msg := range messages {
		messagesHTML.WriteString(renderMessage(msg, cfg))
	}

	navHTML := buildNavHTML(cfg.PrevSessionURL, cfg.NextSessionURL)

	result := template.HTMLTemplate
	result = strings.ReplaceAll(result, "TITLE_PLACEHOLDER", html.EscapeString(cfg.Title))
	result = strings.ReplaceAll(result, "USERNAME_PLACEHOLDER", html.EscapeString(cfg.Username))
	result = strings.ReplaceAll(result, "INITIALS_PLACEHOLDER", html.EscapeString(cfg.UserInitials))
	result = strings.ReplaceAll(result, "NAV_PLACEHOLDER", navHTML)
	result = strings.ReplaceAll(result, "MESSAGES_PLACEHOLDER", messagesHTML.String())

	return result
}

func buildNavHTML(prevURL, nextURL string) string {
	if prevURL == "" && nextURL == "" {
		return ""
	}

	var nav strings.Builder
	nav.WriteString(`<nav class="session-nav">`)

	if prevURL != "" {
		nav.WriteString(`<a href="` + html.EscapeString(prevURL) + `">← Previous Session</a>`)
	} else {
		nav.WriteString(`<span></span>`)
	}

	if nextURL != "" {
		nav.WriteString(`<a href="` + html.EscapeString(nextURL) + `" class="nav-next">Next Session →</a>`)
	}

	nav.WriteString(`</nav>`)
	return nav.String()
}

func mergeBashMessages(messages []parser.Message) []parser.Message {
	var result []parser.Message
	for i := 0; i < len(messages); i++ {
		msg := messages[i]

		if len(msg.Blocks) == 1 && msg.Blocks[0].Type == "bash_input" {
			cmd := msg.Blocks[0].Content
			var stdout, stderr string

			if i+1 < len(messages) && len(messages[i+1].Blocks) == 1 && messages[i+1].Blocks[0].Type == "bash_output" {
				stdout = messages[i+1].Blocks[0].Content
				stderr = messages[i+1].Blocks[0].ToolInput
				i++
			}

			msg.Blocks = []parser.ContentBlock{{
				Type:      "bash_combined",
				Content:   cmd,
				ToolInput: stdout,
				ToolName:  stderr,
			}}
		}

		if len(msg.Blocks) == 1 && msg.Blocks[0].Type == "bash_output" {
			continue
		}

		result = append(result, msg)
	}
	return result
}

func mergeToolResults(messages []parser.Message) []parser.Message {
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

	var result []parser.Message
	for i := 0; i < len(messages); i++ {
		msg := messages[i]

		if msg.Role == "user" && len(msg.Blocks) > 0 && msg.Blocks[0].Type == "tool_result" {
			if len(result) > 0 && result[len(result)-1].Role == "assistant" {
				lastAssistant := &result[len(result)-1]

				for _, resultBlock := range msg.Blocks {
					if resultBlock.ToolUseID != "" {
						resultBlock.ToolName = toolNames[resultBlock.ToolUseID]
						resultBlock.ToolInput = toolInputs[resultBlock.ToolUseID]
					}

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
		if content == "[Request interrupted by user for tool use]" {
			return ""
		}
		if strings.Contains(content, "<thinking>") {
			return renderTextWithThinking(content)
		}
		return `<div class="text-block">` + formatText(block.Content) + `</div>`

	case "thinking":
		return renderThinkingBlock(block.Content)

	case "tool_use":
		return renderToolUse(block.Content)

	case "tool_result":
		return renderToolResult(block)

	case "bash_combined":
		return renderBashCombined(block.Content, block.ToolInput, block.ToolName)

	case "command":
		return renderCommand(block.Content, block.ToolName)

	case "local_command_output":
		content := strings.TrimSpace(block.Content)
		if content == "" || content == "(no content)" {
			return ""
		}
		return `<div class="local-output">` + html.EscapeString(content) + `</div>`

	case "bash_input", "bash_output":
		return ""
	}

	return ""
}

func renderThinkingBlock(content string) string {
	return `<div class="collapsible">
<div class="collapsible-header"><span class="chevron">▶</span> Thinking</div>
<div class="collapsible-content">` + html.EscapeString(content) + `</div>
</div>`
}

var textThinkingRe = regexp.MustCompile(`(?s)<thinking>\s*(.*?)\s*</thinking>`)

func renderTextWithThinking(content string) string {
	matches := textThinkingRe.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return `<div class="text-block">` + formatText(content) + `</div>`
	}

	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		before := strings.TrimSpace(content[lastEnd:match[0]])
		if before != "" {
			result.WriteString(`<div class="text-block">` + formatText(before) + `</div>`)
		}

		thinking := strings.TrimSpace(content[match[2]:match[3]])
		if thinking != "" {
			result.WriteString(renderThinkingBlock(thinking))
		}

		lastEnd = match[1]
	}

	after := strings.TrimSpace(content[lastEnd:])
	if after != "" {
		result.WriteString(`<div class="text-block">` + formatText(after) + `</div>`)
	}

	return result.String()
}

func renderBashCombined(cmd, stdout, stderr string) string {
	termIcon := `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>`

	var result strings.Builder
	result.WriteString(`<div class="tool-block">`)
	result.WriteString(`<div class="tool-pill">` + termIcon + ` Terminal</div>`)
	result.WriteString(`<div class="bash-command"><code>` + html.EscapeString(cmd) + `</code></div>`)

	if stdout != "" || stderr != "" {
		result.WriteString(`<div class="collapsible tool-result">`)
		result.WriteString(`<div class="collapsible-header"><span class="chevron">▶</span> Output</div>`)
		result.WriteString(`<div class="collapsible-content"><pre>`)
		if stdout != "" {
			result.WriteString(html.EscapeString(stdout))
		}
		if stderr != "" {
			result.WriteString(html.EscapeString(stderr))
		}
		result.WriteString(`</pre></div></div>`)
	}

	result.WriteString(`</div>`)
	return result.String()
}

func renderCommand(cmdMsg, cmdName string) string {
	if cmdName == "" {
		return ""
	}
	if strings.HasPrefix(cmdName, "/") {
		return `<div class="slash-command">` + html.EscapeString(cmdName) + `</div>`
	}
	cmdIcon := `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><rect x="8" y="2" width="8" height="4" rx="1" ry="1"/></svg>`
	return `<div class="tool-block command-block">
<div class="tool-pill">` + cmdIcon + ` ` + html.EscapeString(cmdName) + `</div>
</div>`
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

	if toolName == "TodoWrite" {
		return renderTodoWriteTool(toolName, toolInput, icon)
	}

	if toolName == "Task" {
		return renderTaskTool(toolName, toolInput, icon)
	}

	if toolName == "EnterPlanMode" {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` Entering Plan Mode</div></div>`
	}

	if toolName == "ExitPlanMode" {
		return renderExitPlanModeTool(toolName, toolInput, icon)
	}

	if toolName == "AskUserQuestion" {
		return renderAskUserQuestionTool(toolName, toolInput, icon)
	}

	if toolName == "Write" {
		return renderWriteTool(toolName, toolInput, icon)
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

	displayPath := getRelativePath(filePath)

	var result strings.Builder
	result.WriteString(`<div class="tool-block">`)
	result.WriteString(`<div class="tool-pill" title="` + html.EscapeString(filePath) + `">` + icon + ` ` + html.EscapeString(displayPath) + `</div>`)

	if oldString != "" || newString != "" {
		result.WriteString(`<div class="diff-block">`)

		if oldString != "" {
			oldLines := strings.Split(oldString, "\n")
			for _, line := range oldLines {
				result.WriteString(`<div class="diff-line diff-removed">- ` + html.EscapeString(line) + `</div>`)
			}
		}

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

func renderWriteTool(toolName, input, icon string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	filePath, _ := data["file_path"].(string)
	content, _ := data["content"].(string)

	if filePath == "" {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	displayPath := getRelativePath(filePath)

	var result strings.Builder
	result.WriteString(`<div class="tool-block">`)
	result.WriteString(`<div class="tool-pill" title="` + html.EscapeString(filePath) + `">` + icon + ` ` + html.EscapeString(displayPath) + `</div>`)

	if content != "" {
		lang := detectLanguageFromPath(filePath)
		truncated := content
		if len(truncated) > 3000 {
			truncated = truncated[:3000] + "\n... (truncated)"
		}
		result.WriteString(`<div class="collapsible">`)
		result.WriteString(`<div class="collapsible-header"><span class="chevron">▶</span> File Content</div>`)
		result.WriteString(`<div class="collapsible-content"><pre><code class="language-` + lang + `">` + html.EscapeString(truncated) + `</code></pre></div>`)
		result.WriteString(`</div>`)
	}

	result.WriteString(`</div>`)
	return result.String()
}

func renderExitPlanModeTool(toolName, input, icon string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	plan, _ := data["plan"].(string)

	var result strings.Builder
	result.WriteString(`<div class="tool-block">`)
	result.WriteString(`<div class="tool-pill">` + icon + ` ExitPlanMode</div>`)

	if plan != "" {
		result.WriteString(`<div class="collapsible">`)
		result.WriteString(`<div class="collapsible-header"><span class="chevron">▶</span> Plan</div>`)
		result.WriteString(`<div class="collapsible-content"><div class="text-block">` + formatText(plan) + `</div></div>`)
		result.WriteString(`</div>`)
	}

	result.WriteString(`</div>`)
	return result.String()
}

func renderAskUserQuestionTool(toolName, input, icon string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	questions, ok := data["questions"].([]any)
	if !ok || len(questions) == 0 {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	var result strings.Builder
	result.WriteString(`<div class="tool-block question-block">`)
	result.WriteString(`<div class="tool-pill">` + icon + ` AskUserQuestion</div>`)

	for _, q := range questions {
		qMap, ok := q.(map[string]any)
		if !ok {
			continue
		}

		question, _ := qMap["question"].(string)
		header, _ := qMap["header"].(string)
		options, _ := qMap["options"].([]any)

		result.WriteString(`<div class="question-item">`)
		if header != "" {
			result.WriteString(`<div class="question-header">` + html.EscapeString(header) + `</div>`)
		}
		if question != "" {
			result.WriteString(`<div class="question-text">` + html.EscapeString(question) + `</div>`)
		}
		if len(options) > 0 {
			result.WriteString(`<div class="question-options">`)
			for _, opt := range options {
				optMap, ok := opt.(map[string]any)
				if !ok {
					continue
				}
				label, _ := optMap["label"].(string)
				desc, _ := optMap["description"].(string)
				result.WriteString(`<div class="question-option">`)
				result.WriteString(`<span class="option-label">` + html.EscapeString(label) + `</span>`)
				if desc != "" {
					result.WriteString(`<span class="option-desc">` + html.EscapeString(desc) + `</span>`)
				}
				result.WriteString(`</div>`)
			}
			result.WriteString(`</div>`)
		}
		result.WriteString(`</div>`)
	}

	result.WriteString(`</div>`)
	return result.String()
}

func detectLanguageFromPath(filePath string) string {
	ext := strings.ToLower(filePath)
	if idx := strings.LastIndex(ext, "."); idx >= 0 {
		ext = ext[idx:]
	}

	langMap := map[string]string{
		".go": "go", ".py": "python", ".js": "javascript", ".ts": "typescript",
		".tsx": "typescript", ".jsx": "javascript", ".json": "json", ".yaml": "yaml",
		".yml": "yaml", ".md": "markdown", ".sh": "bash", ".bash": "bash",
		".rs": "rust", ".html": "markup", ".xml": "markup", ".css": "css",
		".sql": "sql", ".rb": "ruby", ".java": "java", ".c": "c", ".cpp": "cpp",
	}

	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "plaintext"
}

func renderToolResult(block parser.ContentBlock) string {
	toolName := block.ToolName
	content := block.Content

	if block.IsError {
		return `<div class="tool-result-error">` + html.EscapeString(content) + `</div>`
	}

	if toolName == "Edit" {
		return ""
	}

	if toolName == "Glob" {
		return renderGlobResult(content)
	}

	if toolName == "Grep" {
		return renderGrepResult(content)
	}

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

	if toolName == "Write" {
		return ""
	}

	if toolName == "ExitPlanMode" {
		return `<div class="tool-result-inline plan-approved">✓ User approved the plan</div>`
	}

	if toolName == "AskUserQuestion" {
		return renderAskUserQuestionResult(content)
	}

	if toolName == "TodoWrite" {
		return ""
	}

	if toolName == "Task" {
		return renderTaskResult(content)
	}

	if toolName == "EnterPlanMode" {
		return ""
	}

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

	paths := strings.Split(strings.TrimSpace(content), "\n")
	if len(paths) == 0 {
		return `<div class="tool-result-inline">No matches found</div>`
	}

	looksLikeFilePaths := true
	for _, path := range paths {
		if strings.Contains(path, ":") && !strings.HasPrefix(path, "/") {
			looksLikeFilePaths = false
			break
		}
	}

	if looksLikeFilePaths && len(paths) <= 20 {
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

	truncated := content
	if len(truncated) > 2000 {
		truncated = truncated[:2000] + "\n... (truncated)"
	}

	return `<div class="collapsible tool-result">
<div class="collapsible-header"><span class="chevron">▶</span> Grep Result</div>
<div class="collapsible-content"><pre>` + html.EscapeString(truncated) + `</pre></div>
</div>`
}

func renderTaskResult(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}

	truncated := content
	if len(truncated) > 3000 {
		truncated = truncated[:3000] + "\n... (truncated)"
	}

	return `<div class="collapsible tool-result">
<div class="collapsible-header"><span class="chevron">▶</span> Agent Result</div>
<div class="collapsible-content"><div class="text-block">` + formatText(truncated) + `</div></div>
</div>`
}

func renderTodoWriteTool(toolName, input, icon string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	todos, ok := data["todos"].([]any)
	if !ok || len(todos) == 0 {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	var result strings.Builder
	result.WriteString(`<div class="tool-block todo-block">`)
	result.WriteString(`<div class="tool-pill">` + icon + ` Todo List</div>`)
	result.WriteString(`<div class="todo-list">`)

	for _, t := range todos {
		tMap, ok := t.(map[string]any)
		if !ok {
			continue
		}
		content, _ := tMap["content"].(string)
		status, _ := tMap["status"].(string)

		statusIcon := "○"
		statusClass := "pending"
		switch status {
		case "completed":
			statusIcon = "✓"
			statusClass = "completed"
		case "in_progress":
			statusIcon = "●"
			statusClass = "in-progress"
		}

		result.WriteString(`<div class="todo-item ` + statusClass + `">`)
		result.WriteString(`<span class="todo-status">` + statusIcon + `</span>`)
		result.WriteString(`<span class="todo-content">` + html.EscapeString(content) + `</span>`)
		result.WriteString(`</div>`)
	}

	result.WriteString(`</div></div>`)
	return result.String()
}

func renderTaskTool(toolName, input, icon string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return `<div class="tool-block"><div class="tool-pill">` + icon + ` ` + html.EscapeString(toolName) + `</div></div>`
	}

	subagentType, _ := data["subagent_type"].(string)
	description, _ := data["description"].(string)
	prompt, _ := data["prompt"].(string)

	var result strings.Builder
	result.WriteString(`<div class="tool-block subagent-block">`)
	result.WriteString(`<div class="subagent-header">`)
	result.WriteString(`<span class="subagent-badge">` + icon + ` Task</span>`)
	result.WriteString(`<span class="subagent-note">(subagent) runs independently, doesn't use main context</span>`)
	result.WriteString(`</div>`)

	pillText := subagentType
	if pillText == "" {
		pillText = "Task"
	}
	if description != "" {
		pillText += ": " + description
	}
	result.WriteString(`<div class="subagent-type">` + html.EscapeString(pillText) + `</div>`)

	if prompt != "" {
		truncated := prompt
		if len(truncated) > 500 {
			truncated = truncated[:500] + "..."
		}
		result.WriteString(`<div class="collapsible">`)
		result.WriteString(`<div class="collapsible-header"><span class="chevron">▶</span> Prompt</div>`)
		result.WriteString(`<div class="collapsible-content"><pre>` + html.EscapeString(truncated) + `</pre></div>`)
		result.WriteString(`</div>`)
	}

	result.WriteString(`</div>`)
	return result.String()
}

func renderAskUserQuestionResult(content string) string {
	if !strings.Contains(content, "User has answered") {
		return `<div class="tool-result-inline">` + html.EscapeString(content) + `</div>`
	}

	var result strings.Builder
	result.WriteString(`<div class="question-result">`)
	result.WriteString(`<div class="question-result-header">User's answers:</div>`)

	answerPart := content
	if idx := strings.Index(content, ":"); idx >= 0 {
		answerPart = content[idx+1:]
	}
	if idx := strings.Index(answerPart, ". You can now"); idx >= 0 {
		answerPart = answerPart[:idx]
	}

	pairs := strings.Split(answerPart, "\", \"")
	for _, pair := range pairs {
		pair = strings.Trim(pair, " \"")
		if eqIdx := strings.Index(pair, "\"=\""); eqIdx >= 0 {
			q := pair[:eqIdx]
			a := pair[eqIdx+3:]
			a = strings.TrimSuffix(a, "\"")
			result.WriteString(`<div class="answer-item">`)
			result.WriteString(`<span class="answer-question">` + html.EscapeString(q) + `</span>`)
			result.WriteString(`<span class="answer-value">` + html.EscapeString(a) + `</span>`)
			result.WriteString(`</div>`)
		}
	}

	result.WriteString(`</div>`)
	return result.String()
}

func stripLineNumbers(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	lineNumPattern := regexp.MustCompile(`^\s*\d+→\t?`)
	for _, line := range lines {
		result = append(result, lineNumPattern.ReplaceAllString(line, ""))
	}
	return strings.Join(result, "\n")
}

func getLanguageFromInput(toolInput string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(toolInput), &data); err != nil {
		return "plaintext"
	}
	filePath, _ := data["file_path"].(string)
	if filePath == "" {
		return "plaintext"
	}

	ext := strings.ToLower(filePath)
	if idx := strings.LastIndex(ext, "."); idx >= 0 {
		ext = ext[idx:]
	}

	langMap := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".jsx":  "javascript",
		".json": "json",
		".yaml": "yaml",
		".yml":  "yaml",
		".md":   "markdown",
		".sh":   "bash",
		".bash": "bash",
		".zsh":  "bash",
		".rs":   "rust",
		".html": "markup",
		".xml":  "markup",
		".css":  "css",
		".sql":  "sql",
		".rb":   "ruby",
		".java": "java",
		".c":    "c",
		".cpp":  "cpp",
		".h":    "c",
		".hpp":  "cpp",
	}

	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "plaintext"
}

func getRelativePath(fullPath string) string {
	if currentProjectPath != "" && strings.HasPrefix(fullPath, currentProjectPath) {
		rel := strings.TrimPrefix(fullPath, currentProjectPath)
		return strings.TrimPrefix(rel, "/")
	}
	parts := strings.Split(fullPath, "/")
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return fullPath
}

func getToolIcon(toolName string) string {
	switch toolName {
	case "WebSearch":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/><path d="M11 8v6M8 11h6"/></svg>`
	case "WebFetch":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>`
	case "Read":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>`
	case "Glob":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/><circle cx="12" cy="13" r="3"/></svg>`
	case "Grep":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/><line x1="8" y1="11" x2="14" y2="11"/></svg>`
	case "Write":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="18" x2="12" y2="12"/><line x1="9" y1="15" x2="15" y2="15"/></svg>`
	case "Edit":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>`
	case "Bash":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>`
	case "TodoWrite":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>`
	case "Task":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="9" cy="7" r="4"/><path d="M3 21v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/><path d="M21 21v-2a4 4 0 0 0-3-3.85"/></svg>`
	case "AskUserQuestion":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>`
	case "EnterPlanMode", "ExitPlanMode":
		return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>`
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
