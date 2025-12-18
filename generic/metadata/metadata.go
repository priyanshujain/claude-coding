package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type Session struct {
	PrevSessionID string    `json:"prev_session_id,omitempty"`
	NextSessionID string    `json:"next_session_id,omitempty"`
	GistID        string    `json:"gist_id,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Metadata struct {
	Sessions map[string]Session `json:"sessions"`
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
	return filepath.Join(homeDir, ".claude", "projects", folder, "workbench-metadata.json"), nil
}

func LoadMetadata(projectPath string) (*Metadata, error) {
	path, err := metadataPath(projectPath)
	if err != nil {
		return &Metadata{Sessions: make(map[string]Session)}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Metadata{Sessions: make(map[string]Session)}, nil
		}
		return &Metadata{Sessions: make(map[string]Session)}, err
	}

	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return &Metadata{Sessions: make(map[string]Session)}, err
	}

	if m.Sessions == nil {
		m.Sessions = make(map[string]Session)
	}
	return &m, nil
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

func WithLock(projectPath string, fn func(*Metadata) error) error {
	path, err := metadataPath(projectPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	lockPath := path + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer lockFile.Close()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

	m, _ := LoadMetadata(projectPath)
	if err := fn(m); err != nil {
		return err
	}
	return SaveMetadata(projectPath, m)
}

func (m *Metadata) GetGistID(sessionID string) string {
	if s, ok := m.Sessions[sessionID]; ok {
		return s.GistID
	}
	return ""
}

func (m *Metadata) SetGistID(sessionID, gistID string) {
	s := m.Sessions[sessionID]
	s.GistID = gistID
	s.UpdatedAt = time.Now()
	m.Sessions[sessionID] = s
}

func (m *Metadata) GetPrevSessionID(sessionID string) string {
	if s, ok := m.Sessions[sessionID]; ok {
		return s.PrevSessionID
	}
	return ""
}

func (m *Metadata) GetNextSessionID(sessionID string) string {
	if s, ok := m.Sessions[sessionID]; ok {
		return s.NextSessionID
	}
	return ""
}

func (m *Metadata) ResolveLatestSession(startSessionID string) (latestSession, prevSession string) {
	if startSessionID == "" {
		return "", ""
	}

	if _, ok := m.Sessions[startSessionID]; !ok {
		return startSessionID, ""
	}

	prev := ""
	current := startSessionID
	for {
		next := m.GetNextSessionID(current)
		if next == "" {
			return current, prev
		}
		prev = current
		current = next
	}
}

func (m *Metadata) LinkSession(latestID, newID string) {
	if latestID != "" {
		latest := m.Sessions[latestID]
		latest.NextSessionID = newID
		latest.UpdatedAt = time.Now()
		m.Sessions[latestID] = latest
	}

	newSession := m.Sessions[newID]
	newSession.PrevSessionID = latestID
	newSession.UpdatedAt = time.Now()
	m.Sessions[newID] = newSession
}
