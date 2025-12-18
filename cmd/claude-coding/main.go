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

	if outputPath == "" {
		outputPath = fmt.Sprintf("./thread-%s.html", time.Now().Format("20060102-150405"))
	}

	var sessionFile string
	if sessionID != "" {
		sessionFile, err = parser.FindSessionFileByID(projectPath, sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error finding session %s: %v\n", sessionID, err)
			os.Exit(1)
		}
	} else {
		sessionFile, err = parser.FindSessionFile(projectPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error finding session: %v\n", err)
			os.Exit(1)
		}
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

	cfg := converter.Config{
		Title:        title,
		Username:     username,
		UserInitials: getInitials(username),
		ProjectPath:  projectPath,
	}

	html := converter.Convert(messages, cfg)

	if createGist {
		gistOK := true
		if !gist.IsGHAvailable() {
			fmt.Fprintf(os.Stderr, "warning: gh CLI is not installed, skipping gist creation\n")
			gistOK = false
		} else if !gist.IsGHAuthenticated() {
			fmt.Fprintf(os.Stderr, "warning: gh CLI is not authenticated, skipping gist creation\n")
			gistOK = false
		}

		if gistOK {
			metadata := gist.LoadMetadata(projectPath)
			filename := "claude-conversation.html"
			if sessionID != "" {
				filename = "claude-code-" + sessionID + ".html"
			}

			var previewURL string
			var gistID string

			existingGistID := ""
			if sessionID != "" {
				existingGistID = metadata.GetGistID(sessionID)
			}

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
				if sessionID != "" && gistID != "" {
					metadata.SetGistID(sessionID, gistID)
					gist.SaveMetadata(projectPath, metadata)
				}
				fmt.Println(previewURL)
				return
			}
		}
	}

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
