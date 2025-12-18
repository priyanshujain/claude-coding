package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

type HookInput struct {
	SessionID string `json:"session_id"`
	Cwd       string `json:"cwd"`
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

	hash := md5.Sum([]byte(hookInput.Cwd))
	prevSessionFile := filepath.Join(os.TempDir(), "claude-prev-session-"+hex.EncodeToString(hash[:]))
	os.WriteFile(prevSessionFile, []byte(hookInput.SessionID), 0644)
}
