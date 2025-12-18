package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/priyanshujain/claude-coding/generic/metadata"
)

type HookInput struct {
	SessionID string `json:"session_id"`
	Source    string `json:"source"`
	Cwd       string `json:"cwd"`
}

func readPrevSessionWithRetry(filePath string, maxRetries int, delay time.Duration) string {
	for i := 0; i < maxRetries; i++ {
		data, err := os.ReadFile(filePath)
		if err == nil && len(data) > 0 {
			return string(data)
		}
		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}
	return ""
}

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}

	var hookInput HookInput
	if err := json.Unmarshal(input, &hookInput); err != nil {
		return
	}

	if hookInput.SessionID == "" || hookInput.Cwd == "" {
		return
	}

	var prevSessionID string
	hash := md5.Sum([]byte(hookInput.Cwd))
	prevSessionFile := filepath.Join(os.TempDir(), "claude-prev-session-"+hex.EncodeToString(hash[:]))

	if hookInput.Source == "clear" {
		prevSessionID = readPrevSessionWithRetry(prevSessionFile, 10, 100*time.Millisecond)
		if prevSessionID != "" {
			os.Remove(prevSessionFile)
		}
	}

	metadata.WithLock(hookInput.Cwd, func(m *metadata.Metadata) error {
		if prevSessionID != "" && prevSessionID != hookInput.SessionID {
			m.LinkSession(prevSessionID, hookInput.SessionID)
		} else {
			m.LinkSession("", hookInput.SessionID)
		}
		return nil
	})

	envFile := os.Getenv("CLAUDE_ENV_FILE")
	if envFile != "" {
		f, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			fmt.Fprintf(f, "export CLAUDE_SESSION_ID='%s'\n", hookInput.SessionID)
			f.Close()
		}
	}
}
