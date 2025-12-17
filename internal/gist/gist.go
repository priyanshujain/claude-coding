package gist

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func IsGHAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func IsGHAuthenticated() bool {
	cmd := exec.Command("gh", "auth", "status")
	return cmd.Run() == nil
}

func Create(htmlContent string) (string, error) {
	if !IsGHAvailable() {
		return "", fmt.Errorf("gh CLI is not installed")
	}

	tmpFile, err := os.CreateTemp("", "claude-thread-*.html")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(htmlContent); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command("gh", "gist", "create", "--public", tmpFile.Name())
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create gist: %w", err)
	}

	gistURL := strings.TrimSpace(string(output))
	gistID := filepath.Base(gistURL)

	previewURL := fmt.Sprintf("https://gistpreview.github.io/?%s", gistID)
	return previewURL, nil
}
