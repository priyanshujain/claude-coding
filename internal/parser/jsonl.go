package parser

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type rawMessage struct {
	Type      string     `json:"type"`
	UUID      string     `json:"uuid"`
	Timestamp string     `json:"timestamp"`
	Message   rawContent `json:"message"`
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
	ID        string // Message ID for grouping streaming chunks
	Role      string
	Timestamp time.Time
	Blocks    []ContentBlock
}

type ContentBlock struct {
	Type      string
	Content   string
	ToolName  string // For tool_result blocks, the name of the tool that produced it
	ToolUseID string // For both tool_use and tool_result
	ToolInput string // Raw JSON input for tool_use blocks
	IsError   bool   // For tool_result blocks, whether the tool call failed/was rejected
}

func FindSessionFile(projectPath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	projectFolder := strings.ReplaceAll(projectPath, "/", "-")
	projectFolder = strings.ReplaceAll(projectFolder, ".", "-")

	claudeProjectDir := filepath.Join(homeDir, ".claude", "projects", projectFolder)

	entries, err := os.ReadDir(claudeProjectDir)
	if err != nil {
		return "", err
	}

	var latestFile string
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".jsonl") || strings.HasPrefix(name, "agent-") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestFile = filepath.Join(claudeProjectDir, name)
		}
	}

	if latestFile == "" {
		return "", os.ErrNotExist
	}

	return latestFile, nil
}

// FindSessionFileByID finds a session file by its session ID
func FindSessionFileByID(projectPath, sessionID string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	projectFolder := strings.ReplaceAll(projectPath, "/", "-")
	projectFolder = strings.ReplaceAll(projectFolder, ".", "-")

	claudeProjectDir := filepath.Join(homeDir, ".claude", "projects", projectFolder)
	sessionFile := filepath.Join(claudeProjectDir, sessionID+".jsonl")

	if _, err := os.Stat(sessionFile); err != nil {
		return "", err
	}

	return sessionFile, nil
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

		msg := parseRawMessage(raw)
		if len(msg.Blocks) > 0 {
			// Merge with previous message if same ID (streaming chunks)
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
		msg.Blocks = append(msg.Blocks, ContentBlock{
			Type:    "text",
			Content: content,
		})
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
		content, _ := data["content"].(string)
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
