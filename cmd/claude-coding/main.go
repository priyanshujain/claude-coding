package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/priyanshujain/claude-coding/internal/converter"
	"github.com/priyanshujain/claude-coding/internal/gist"
	"github.com/priyanshujain/claude-coding/generic/metadata"
	"github.com/priyanshujain/claude-coding/internal/parser"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "share":
		shareCmd(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: claude-coding <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  share    Export conversation thread to HTML")
	fmt.Println()
	fmt.Println("Run 'claude-coding <command> -h' for command-specific help")
}

func shareCmd(args []string) {
	fs := flag.NewFlagSet("share", flag.ExitOnError)

	var projectPath string
	var outputPath string
	var title string
	var username string
	var sessionID string
	var createGist bool

	fs.StringVar(&projectPath, "project", "", "project path")
	fs.StringVar(&outputPath, "output", "", "output file path")
	fs.StringVar(&title, "title", "", "thread title")
	fs.StringVar(&username, "username", "", "username to display")
	fs.StringVar(&sessionID, "session", "", "specific session ID to export")
	fs.BoolVar(&createGist, "gist", false, "create GitHub gist and return preview URL")
	fs.Parse(args)

	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		projectPath = cwd
	}

	projectPath, err := filepath.Abs(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	sessionID, err = parser.ResolveCurrentSessionID(projectPath, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error finding session: %v\n", err)
		os.Exit(1)
	}

	m, _ := metadata.LoadMetadata(projectPath)
	prevSessionID := m.GetPrevSessionID(sessionID)
	nextSessionID := m.GetNextSessionID(sessionID)

	if outputPath == "" {
		outputPath = fmt.Sprintf("./thread-%s.html", time.Now().Format("20060102-150405"))
	}

	sessionFile, err := parser.GetSessionFilePath(projectPath, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error finding session file: %v\n", err)
		os.Exit(1)
	}

	messages, err := parser.ParseFile(sessionFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing session: %v\n", err)
		os.Exit(1)
	}

	if len(messages) == 0 {
		fmt.Fprintf(os.Stderr, "error: no messages found in session\n")
		os.Exit(1)
	}

	if title == "" {
		title = parser.ParseSummary(sessionFile)
	}
	if title == "" {
		title = extractTitle(messages)
	}

	if username == "" {
		username = getSystemUsername()
	}

	if createGist {
		gistOK := true
		if !gist.IsGHAvailable() {
			fmt.Fprintf(os.Stderr, "warning: gh CLI is not installed, skipping gist creation\n")
			gistOK = false
		} else if !gist.IsGHAuthenticated() {
			fmt.Fprintf(os.Stderr, "warning: gh CLI is not authenticated, skipping gist creation\n")
			gistOK = false
		}

		if gistOK && sessionID != "" {
			m, _ := metadata.LoadMetadata(projectPath)

			ensureSessionGist(projectPath, m, prevSessionID)
			ensureSessionGist(projectPath, m, nextSessionID)

			m, _ = metadata.LoadMetadata(projectPath)

			var prevSessionURL, nextSessionURL string
			if prevSessionID != "" {
				if prevGistID := m.GetGistID(prevSessionID); prevGistID != "" {
					prevSessionURL = gist.PreviewURL(prevGistID)
				}
			}
			if nextSessionID != "" {
				if nextGistID := m.GetGistID(nextSessionID); nextGistID != "" {
					nextSessionURL = gist.PreviewURL(nextGistID)
				}
			}

			cfg := converter.Config{
				Title:          title,
				Username:       username,
				UserInitials:   getInitials(username),
				ProjectPath:    projectPath,
				PrevSessionURL: prevSessionURL,
				NextSessionURL: nextSessionURL,
			}
			html := converter.Convert(messages, cfg)

			filename := "claude-code-" + sessionID + ".html"
			var previewURL string
			var gistID string

			existingGistID := m.GetGistID(sessionID)
			if existingGistID != "" {
				previewURL, err = gist.Update(existingGistID, filename, html)
				if err != nil {
					gistID, previewURL, err = gist.Create(filename, html)
				} else {
					gistID = existingGistID
				}
			} else {
				gistID, previewURL, err = gist.Create(filename, html)
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to create/update gist: %v\n", err)
			} else {
				m.SetGistID(sessionID, gistID)
				metadata.SaveMetadata(projectPath, m)
				updateSessionChain(projectPath, m, sessionID, previewURL)
				fmt.Println(previewURL)
				return
			}
		}
	}

	var prevSessionURL, nextSessionURL string
	if prevSessionID != "" || nextSessionID != "" {
		m, _ := metadata.LoadMetadata(projectPath)
		if prevSessionID != "" {
			if prevGistID := m.GetGistID(prevSessionID); prevGistID != "" {
				prevSessionURL = gist.PreviewURL(prevGistID)
			}
		}
		if nextSessionID != "" {
			if nextGistID := m.GetGistID(nextSessionID); nextGistID != "" {
				nextSessionURL = gist.PreviewURL(nextGistID)
			}
		}
	}

	cfg := converter.Config{
		Title:          title,
		Username:       username,
		UserInitials:   getInitials(username),
		ProjectPath:    projectPath,
		PrevSessionURL: prevSessionURL,
		NextSessionURL: nextSessionURL,
	}
	html := converter.Convert(messages, cfg)

	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}

	absOutput, _ := filepath.Abs(outputPath)
	fmt.Printf("Thread exported to: %s\n", absOutput)
}

func extractTitle(messages []parser.Message) string {
	for _, msg := range messages {
		if msg.Role == "user" {
			for _, block := range msg.Blocks {
				if block.Type == "text" && block.Content != "" {
					title := strings.TrimSpace(block.Content)
					if strings.HasPrefix(title, "Caveat:") || strings.HasPrefix(title, "<") {
						continue
					}
					if len(title) > 80 {
						title = title[:77] + "..."
					}
					lines := strings.Split(title, "\n")
					return lines[0]
				}
			}
		}
	}
	return "Claude Code Thread"
}

func getSystemUsername() string {
	if u, err := user.Current(); err == nil {
		if u.Name != "" {
			return u.Name
		}
		return u.Username
	}
	return "User"
}

func getInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "U"
	}
	if len(parts) == 1 {
		if len(parts[0]) >= 2 {
			return strings.ToUpper(parts[0][:2])
		}
		return strings.ToUpper(parts[0][:1])
	}
	return strings.ToUpper(string(parts[0][0]) + string(parts[len(parts)-1][0]))
}

func ensureSessionGist(projectPath string, m *metadata.Metadata, sessionID string) {
	if sessionID == "" {
		return
	}
	if m.GetGistID(sessionID) != "" {
		return
	}

	sessionFile, err := parser.GetSessionFilePath(projectPath, sessionID)
	if err != nil {
		return
	}

	messages, err := parser.ParseFile(sessionFile)
	if err != nil || len(messages) == 0 {
		return
	}

	title := parser.ParseSummary(sessionFile)
	if title == "" {
		title = extractTitle(messages)
	}

	cfg := converter.Config{
		Title:        title,
		Username:     getSystemUsername(),
		UserInitials: getInitials(getSystemUsername()),
		ProjectPath:  projectPath,
	}

	html := converter.Convert(messages, cfg)
	filename := "claude-code-" + sessionID + ".html"

	gistID, _, err := gist.Create(filename, html)
	if err != nil {
		return
	}
	m.SetGistID(sessionID, gistID)
	metadata.SaveMetadata(projectPath, m)
}

func updateSessionChain(projectPath string, m *metadata.Metadata, currentSessionID, currentGistURL string) {
	syncSessionGist(projectPath, m, m.GetPrevSessionID(currentSessionID), currentGistURL, "next")
	syncSessionGist(projectPath, m, m.GetNextSessionID(currentSessionID), currentGistURL, "prev")
}

func syncSessionGist(projectPath string, m *metadata.Metadata, sessionID, adjacentURL, direction string) {
	if sessionID == "" {
		return
	}

	sessionFile, err := parser.GetSessionFilePath(projectPath, sessionID)
	if err != nil {
		return
	}

	messages, err := parser.ParseFile(sessionFile)
	if err != nil || len(messages) == 0 {
		return
	}

	var prevURL, nextURL string
	if direction == "next" {
		nextURL = adjacentURL
		if prevID := m.GetPrevSessionID(sessionID); prevID != "" {
			if prevGistID := m.GetGistID(prevID); prevGistID != "" {
				prevURL = gist.PreviewURL(prevGistID)
			}
		}
	} else {
		prevURL = adjacentURL
		if nextID := m.GetNextSessionID(sessionID); nextID != "" {
			if nextGistID := m.GetGistID(nextID); nextGistID != "" {
				nextURL = gist.PreviewURL(nextGistID)
			}
		}
	}

	title := parser.ParseSummary(sessionFile)
	if title == "" {
		title = extractTitle(messages)
	}

	cfg := converter.Config{
		Title:          title,
		Username:       getSystemUsername(),
		UserInitials:   getInitials(getSystemUsername()),
		ProjectPath:    projectPath,
		PrevSessionURL: prevURL,
		NextSessionURL: nextURL,
	}

	html := converter.Convert(messages, cfg)
	filename := "claude-code-" + sessionID + ".html"

	existingGistID := m.GetGistID(sessionID)
	var gistID string

	if existingGistID != "" {
		_, err = gist.Update(existingGistID, filename, html)
		if err != nil {
			gistID, _, err = gist.Create(filename, html)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to create %s session gist: %v\n", direction, err)
				return
			}
			m.SetGistID(sessionID, gistID)
			metadata.SaveMetadata(projectPath, m)
		} else {
			gistID = existingGistID
		}
	} else {
		gistID, _, err = gist.Create(filename, html)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to create %s session gist: %v\n", direction, err)
			return
		}
		m.SetGistID(sessionID, gistID)
		metadata.SaveMetadata(projectPath, m)
	}

	newURL := gist.PreviewURL(gistID)
	if direction == "next" {
		syncSessionGist(projectPath, m, m.GetPrevSessionID(sessionID), newURL, "next")
	} else {
		syncSessionGist(projectPath, m, m.GetNextSessionID(sessionID), newURL, "prev")
	}
}
