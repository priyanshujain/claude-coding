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

func buildDescription(gistID string) string {
	return fmt.Sprintf("Claude Code conversation export\nPreview: https://gistpreview.github.io/?%s\n⚠️ Do not delete - shared preview link will break", gistID)
}

func PreviewURL(gistID string) string {
	return fmt.Sprintf("https://gistpreview.github.io/?%s", gistID)
}

func Create(filename, htmlContent string) (gistID string, previewURL string, err error) {
	if !IsGHAvailable() {
		return "", "", fmt.Errorf("gh CLI is not installed")
	}

	tmpDir, err := os.MkdirTemp("", "claude-gist-*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFilePath := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(tmpFilePath, []byte(htmlContent), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write temp file: %w", err)
	}

	cmd := exec.Command("gh", "gist", "create", "--public", tmpFilePath)
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to create gist: %w", err)
	}

	gistURL := strings.TrimSpace(string(output))
	gistID = filepath.Base(gistURL)
	previewURL = PreviewURL(gistID)

	desc := buildDescription(gistID)
	descCmd := exec.Command("gh", "gist", "edit", gistID, "--desc", desc)
	descCmd.Run()

	return gistID, previewURL, nil
}

func Update(gistID, filename, htmlContent string) (previewURL string, err error) {
	if !IsGHAvailable() {
		return "", fmt.Errorf("gh CLI is not installed")
	}

	tmpDir, err := os.MkdirTemp("", "claude-gist-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFilePath := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(tmpFilePath, []byte(htmlContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	desc := buildDescription(gistID)
	cmd := exec.Command("gh", "gist", "edit", gistID, "--add", tmpFilePath, "--desc", desc)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to update gist: %w", err)
	}

	return PreviewURL(gistID), nil
}
