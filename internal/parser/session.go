package parser

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/priyanshujain/claude-coding/generic/metadata"
)

func FindLatestSessionID(projectPath string) (string, error) {
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

	var latestID string
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
			latestID = strings.TrimSuffix(name, ".jsonl")
		}
	}

	if latestID == "" {
		return "", os.ErrNotExist
	}
	return latestID, nil
}

func ResolveCurrentSessionID(projectPath, sessionID string) (string, error) {
	if sessionID == "" {
		var err error
		sessionID, err = FindLatestSessionID(projectPath)
		if err != nil {
			return "", err
		}
	}

	m, _ := metadata.LoadMetadata(projectPath)
	resolved, _ := m.ResolveLatestSession(sessionID)
	if resolved != "" {
		sessionID = resolved
	}

	return sessionID, nil
}

func GetSessionFilePath(projectPath, sessionID string) (string, error) {
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
