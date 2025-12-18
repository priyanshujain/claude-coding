package gist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SessionGist struct {
	GistID    string    `json:"gist_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Metadata struct {
	Sessions map[string]SessionGist `json:"sessions"`
}

func encodeProjectPath(projectPath string) string {
	folder := strings.ReplaceAll(projectPath, "/", "-")
	folder = strings.ReplaceAll(folder, ".", "-")
	return folder
}

func metadataPath(projectPath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	folder := encodeProjectPath(projectPath)
	return filepath.Join(homeDir, ".claude", "projects", folder, "gist-metadata.json"), nil
}

func LoadMetadata(projectPath string) *Metadata {
	path, err := metadataPath(projectPath)
	if err != nil {
		return &Metadata{Sessions: make(map[string]SessionGist)}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return &Metadata{Sessions: make(map[string]SessionGist)}
	}

	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return &Metadata{Sessions: make(map[string]SessionGist)}
	}

	if m.Sessions == nil {
		m.Sessions = make(map[string]SessionGist)
	}
	return &m
}

func SaveMetadata(projectPath string, m *Metadata) error {
	path, err := metadataPath(projectPath)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (m *Metadata) GetGistID(sessionID string) string {
	if sg, ok := m.Sessions[sessionID]; ok {
		return sg.GistID
	}
	return ""
}

func (m *Metadata) SetGistID(sessionID, gistID string) {
	m.Sessions[sessionID] = SessionGist{
		GistID:    gistID,
		UpdatedAt: time.Now(),
	}
}
