package parser

import (
	"bufio"
	"encoding/json"
	"os"
	"regexp"
	"time"
)

type rawMessage struct {
	Type      string     `json:"type"`
	UUID      string     `json:"uuid"`
	Timestamp string     `json:"timestamp"`
	Message   rawContent `json:"message"`
	IsMeta    bool       `json:"isMeta"`
	Summary   string     `json:"summary"`
}

type rawContent struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type rawContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
	Name     string `json:"name,omitempty"`
	Input    any    `json:"input,omitempty"`
	Content  string `json:"content,omitempty"`
}

type Message struct {
	ID        string
	Role      string
	Timestamp time.Time
	Blocks    []ContentBlock
}

type ContentBlock struct {
	Type      string
	Content   string
	ToolName  string
	ToolUseID string
	ToolInput string
	IsError   bool
}

var (
	bashInputRe       = regexp.MustCompile(`<bash-input>([\s\S]*?)</bash-input>`)
	bashStdoutRe      = regexp.MustCompile(`<bash-stdout>([\s\S]*?)</bash-stdout>`)
	bashStderrRe      = regexp.MustCompile(`<bash-stderr>([\s\S]*?)</bash-stderr>`)
	commandMsgRe      = regexp.MustCompile(`<command-message>([\s\S]*?)</command-message>`)
	commandNameRe     = regexp.MustCompile(`<command-name>([\s\S]*?)</command-name>`)
	localCmdStdoutRe  = regexp.MustCompile(`<local-command-stdout>([\s\S]*?)</local-command-stdout>`)
)

func ParseSummary(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	var lastSummary string
	for scanner.Scan() {
		var raw rawMessage
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}
		if raw.Type == "summary" && raw.Summary != "" {
			lastSummary = raw.Summary
		}
	}
	return lastSummary
}

func ParseFile(filePath string) ([]Message, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var messages []Message
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		var raw rawMessage
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}

		if raw.Type != "user" && raw.Type != "assistant" {
			continue
		}

		if raw.IsMeta {
			continue
		}

		msg := parseRawMessage(raw)
		if len(msg.Blocks) > 0 {
			if len(messages) > 0 && msg.ID != "" && messages[len(messages)-1].ID == msg.ID {
				messages[len(messages)-1].Blocks = append(messages[len(messages)-1].Blocks, msg.Blocks...)
			} else {
				messages = append(messages, msg)
			}
		}
	}

	return messages, scanner.Err()
}

func parseRawMessage(raw rawMessage) Message {
	msg := Message{
		ID:   raw.Message.ID,
		Role: raw.Message.Role,
	}

	if t, err := time.Parse(time.RFC3339, raw.Timestamp); err == nil {
		msg.Timestamp = t
	}

	switch content := raw.Message.Content.(type) {
	case string:
		blocks := parseSpecialContent(content)
		msg.Blocks = append(msg.Blocks, blocks...)
	case []any:
		for _, item := range content {
			block := parseContentBlock(item)
			if block.Type != "" {
				msg.Blocks = append(msg.Blocks, block)
			}
		}
	}

	return msg
}

func parseSpecialContent(content string) []ContentBlock {
	if matches := bashInputRe.FindStringSubmatch(content); len(matches) > 1 {
		return []ContentBlock{{Type: "bash_input", Content: matches[1]}}
	}

	if bashStdoutRe.MatchString(content) || bashStderrRe.MatchString(content) {
		var stdout, stderr string
		if matches := bashStdoutRe.FindStringSubmatch(content); len(matches) > 1 {
			stdout = matches[1]
		}
		if matches := bashStderrRe.FindStringSubmatch(content); len(matches) > 1 {
			stderr = matches[1]
		}
		return []ContentBlock{{Type: "bash_output", Content: stdout, ToolInput: stderr}}
	}

	if commandMsgRe.MatchString(content) {
		var cmdMsg, cmdName string
		if matches := commandMsgRe.FindStringSubmatch(content); len(matches) > 1 {
			cmdMsg = matches[1]
		}
		if matches := commandNameRe.FindStringSubmatch(content); len(matches) > 1 {
			cmdName = matches[1]
		}
		return []ContentBlock{{Type: "command", Content: cmdMsg, ToolName: cmdName}}
	}

	if matches := localCmdStdoutRe.FindStringSubmatch(content); len(matches) > 1 {
		return []ContentBlock{{Type: "local_command_output", Content: matches[1]}}
	}

	return []ContentBlock{{Type: "text", Content: content}}
}

func parseContentBlock(item any) ContentBlock {
	data, ok := item.(map[string]any)
	if !ok {
		return ContentBlock{}
	}

	blockType, _ := data["type"].(string)

	switch blockType {
	case "text":
		text, _ := data["text"].(string)
		return ContentBlock{Type: "text", Content: text}

	case "thinking":
		thinking, _ := data["thinking"].(string)
		return ContentBlock{Type: "thinking", Content: thinking}

	case "tool_use":
		name, _ := data["name"].(string)
		id, _ := data["id"].(string)
		input, _ := json.MarshalIndent(data["input"], "", "  ")
		return ContentBlock{
			Type:      "tool_use",
			Content:   name + "\n" + string(input),
			ToolName:  name,
			ToolUseID: id,
			ToolInput: string(input),
		}

	case "tool_result":
		var content string
		switch c := data["content"].(type) {
		case string:
			content = c
		case []any:
			for _, item := range c {
				if textObj, ok := item.(map[string]any); ok {
					if text, ok := textObj["text"].(string); ok {
						if content != "" {
							content += "\n"
						}
						content += text
					}
				}
			}
		}
		toolUseID, _ := data["tool_use_id"].(string)
		isError, _ := data["is_error"].(bool)
		return ContentBlock{
			Type:      "tool_result",
			Content:   content,
			ToolUseID: toolUseID,
			IsError:   isError,
		}
	}

	return ContentBlock{}
}
