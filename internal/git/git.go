package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// GetRepoName returns the repository name from git remote origin URL.
func GetRepoName() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}

	url := strings.TrimSpace(out.String())
	if url == "" {
		return ""
	}

	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) == 2 {
			url = parts[1]
		}
	}

	url = strings.TrimSuffix(url, ".git")

	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}

func CopyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "windows":
		cmd = exec.Command("clip")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func OpenInBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// OpenInLinear opens a Linear issue URL in the Linear desktop app.
// On macOS, it first checks if Linear.app is installed and uses it directly.
// If the app is not installed, it falls back to opening in the browser.
func OpenInLinear(url string) error {
	switch runtime.GOOS {
	case "darwin":
		// Check if Linear.app is installed
		if _, err := os.Stat("/Applications/Linear.app"); err == nil {
			// Linear app is installed, open URL with it directly
			cmd := exec.Command("open", "-a", "Linear", url)
			return cmd.Start()
		}
		// Fall back to browser
		return OpenInBrowser(url)
	case "linux", "windows":
		// On other platforms, just open in browser
		return OpenInBrowser(url)
	default:
		return OpenInBrowser(url)
	}
}

type TerminalConfig struct {
	Terminal string
	Command  string
}

func OpenTerminalWithOpencode(workDir string, inputCommand string, cfg TerminalConfig) error {
	terminal := cfg.Terminal
	if terminal == "" || terminal == "auto" {
		terminal = detectTerminal()
	}

	opencodeCmd := cfg.Command
	if opencodeCmd == "" {
		opencodeCmd = "opencode"
	}

	cmd, err := buildTerminalCommand(terminal, workDir, opencodeCmd, inputCommand)
	if err != nil {
		return err
	}

	return cmd.Start()
}

func detectTerminal() string {
	if os.Getenv("TMUX") != "" {
		return "tmux"
	}
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return "kitty"
	}
	if os.Getenv("WEZTERM_PANE") != "" {
		return "wezterm"
	}

	termProgram := os.Getenv("TERM_PROGRAM")
	switch termProgram {
	case "ghostty":
		return "ghostty"
	case "iTerm.app":
		return "iterm"
	case "Apple_Terminal":
		return "terminal"
	}

	if os.Getenv("TERM") == "xterm-ghostty" {
		return "ghostty"
	}

	switch runtime.GOOS {
	case "darwin":
		return "terminal"
	case "linux":
		return "gnome-terminal"
	default:
		return "terminal"
	}
}

func buildTerminalCommand(terminal, workDir, opencodeCmd, inputCommand string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "darwin":
		return buildDarwinCommand(terminal, workDir, opencodeCmd, inputCommand)
	case "linux":
		return buildLinuxCommand(terminal, workDir, opencodeCmd, inputCommand)
	case "windows":
		return buildWindowsCommand(workDir, opencodeCmd, inputCommand)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func buildDarwinCommand(terminal, workDir, opencodeCmd, inputCommand string) (*exec.Cmd, error) {
	switch terminal {
	case "ghostty":
		// Use ghostty CLI directly with -e flag to run command in new window
		return exec.Command("ghostty",
			"--working-directory="+workDir,
			"-e", "sh", "-c",
			fmt.Sprintf("%s --prompt %q", opencodeCmd, inputCommand),
		), nil

	case "iterm":
		script := fmt.Sprintf(`
tell application "iTerm"
    set newWindow to (create window with default profile)
    tell current session of newWindow
        write text "cd %s && %s --prompt %s"
    end tell
end tell`, escapeForAppleScript(workDir), opencodeCmd, escapeForAppleScript(inputCommand))
		return exec.Command("osascript", "-e", script), nil

	case "tmux":
		return exec.Command("sh", "-c", fmt.Sprintf(
			`tmux new-window -c %q '%s --prompt %q'`,
			workDir, opencodeCmd, inputCommand,
		)), nil

	default:
		script := fmt.Sprintf(`
tell application "Terminal"
    do script "cd %s && %s --prompt %s"
    activate
end tell`, escapeForAppleScript(workDir), opencodeCmd, escapeForAppleScript(inputCommand))
		return exec.Command("osascript", "-e", script), nil
	}
}

func buildLinuxCommand(terminal, workDir, opencodeCmd, inputCommand string) (*exec.Cmd, error) {
	switch terminal {
	case "ghostty":
		return exec.Command("sh", "-c", fmt.Sprintf(
			`ghostty +new-window -e '%s --prompt %q'`,
			opencodeCmd, inputCommand,
		)), nil

	case "kitty":
		return exec.Command("sh", "-c", fmt.Sprintf(
			`kitty @ launch --type=os-window --cwd=%q %s --prompt %q`,
			workDir, opencodeCmd, inputCommand,
		)), nil

	case "wezterm":
		return exec.Command("sh", "-c", fmt.Sprintf(
			`wezterm cli spawn --new-window --cwd %q -- %s --prompt %q`,
			workDir, opencodeCmd, inputCommand,
		)), nil

	case "tmux":
		return exec.Command("sh", "-c", fmt.Sprintf(
			`tmux new-window -c %q '%s --prompt %q'`,
			workDir, opencodeCmd, inputCommand,
		)), nil

	default:
		return exec.Command("sh", "-c", fmt.Sprintf(
			`gnome-terminal --window --working-directory=%q -- %s --prompt %q`,
			workDir, opencodeCmd, inputCommand,
		)), nil
	}
}

func buildWindowsCommand(workDir, opencodeCmd, inputCommand string) (*exec.Cmd, error) {
	shellCmd := fmt.Sprintf(
		`cd /d "%s" && %s --prompt "%s"`,
		workDir, opencodeCmd, inputCommand,
	)
	return exec.Command("cmd", "/c", "start", "cmd", "/k", shellCmd), nil
}

func escapeShellString(s string) string {
	return strings.ReplaceAll(s, "'", "'\"'\"'")
}

func escapeForAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
