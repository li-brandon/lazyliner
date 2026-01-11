package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/brandonli/lazyliner/internal/app"
	"github.com/brandonli/lazyliner/internal/config"
	"github.com/brandonli/lazyliner/internal/linear"
	"github.com/brandonli/lazyliner/internal/ui/theme"
	"github.com/brandonli/lazyliner/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "lazyliner",
	Short: "A terminal TUI for Linear",
	Long:  "Lazyliner is a beautiful, keyboard-driven terminal interface for Linear issue tracking.",
	RunE:  runTUI,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues",
	RunE:  runList,
}

var viewCmd = &cobra.Command{
	Use:   "view [id]",
	Short: "View issue details",
	Args:  cobra.ExactArgs(1),
	RunE:  runView,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new issue",
	RunE:  runCreate,
}

var (
	listLimit int
	listMine  bool
)

func init() {
	listCmd.Flags().IntVarP(&listLimit, "limit", "n", 20, "Number of issues to display")
	listCmd.Flags().BoolVarP(&listMine, "mine", "m", false, "Show only my issues")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(createCmd)
}

func main() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func requireAPIKey() error {
	if cfg.Linear.APIKey == "" {
		fmt.Fprintln(os.Stderr, "Linear API key not configured.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Set it via environment variable:")
		fmt.Fprintln(os.Stderr, "  export LAZYLINER_API_KEY=lin_api_xxxxx")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Or create a config file at ~/.config/lazyliner/config.yaml:")
		fmt.Fprintln(os.Stderr, "  linear:")
		fmt.Fprintln(os.Stderr, "    api_key: lin_api_xxxxx")
		return fmt.Errorf("API key not configured")
	}
	return nil
}

func runTUI(cmd *cobra.Command, args []string) error {
	// Don't require API key - the TUI will show setup instructions if not configured
	p := tea.NewProgram(
		app.New(cfg),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running app: %w", err)
	}
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	if err := requireAPIKey(); err != nil {
		return err
	}

	client := linear.NewClient(cfg.Linear.APIKey)
	ctx := context.Background()

	var issues []linear.Issue
	var err error

	if listMine {
		issues, err = client.GetMyIssues(ctx, listLimit)
	} else {
		issues, err = client.GetIssues(ctx, linear.IssueFilter{Limit: listLimit})
	}

	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tPRIORITY\tASSIGNEE")
	fmt.Fprintln(w, "──\t─────\t──────\t────────\t────────")

	for _, issue := range issues {
		status := "Unknown"
		if issue.State != nil {
			status = issue.State.Name
		}
		assignee := "-"
		if issue.Assignee != nil {
			assignee = issue.Assignee.Name
		}
		priority := theme.PriorityLabel(issue.Priority)

		title := util.Truncate(issue.Title, 50)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			issue.Identifier,
			title,
			status,
			priority,
			assignee,
		)
	}
	w.Flush()

	fmt.Printf("\nShowing %d issues\n", len(issues))
	return nil
}

func runView(cmd *cobra.Command, args []string) error {
	if err := requireAPIKey(); err != nil {
		return err
	}

	client := linear.NewClient(cfg.Linear.APIKey)
	ctx := context.Background()

	issue, err := client.GetIssue(ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to fetch issue: %w", err)
	}

	status := "Unknown"
	if issue.State != nil {
		status = issue.State.Name
	}
	assignee := "Unassigned"
	if issue.Assignee != nil {
		assignee = issue.Assignee.Name
	}

	fmt.Printf("╭────────────────────────────────────────────────────────────╮\n")
	fmt.Printf("│ %s: %s\n", issue.Identifier, util.Truncate(issue.Title, 50))
	fmt.Printf("├────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│ Status:   %s\n", status)
	fmt.Printf("│ Priority: %s\n", theme.PriorityLabel(issue.Priority))
	fmt.Printf("│ Assignee: %s\n", assignee)
	fmt.Printf("│ URL:      %s\n", issue.URL)
	if issue.BranchName != "" {
		fmt.Printf("│ Branch:   %s\n", issue.BranchName)
	}
	fmt.Printf("├────────────────────────────────────────────────────────────┤\n")

	if issue.Description != "" {
		fmt.Printf("│ Description:\n")
		for _, line := range strings.Split(issue.Description, "\n") {
			fmt.Printf("│   %s\n", util.Truncate(line, 56))
		}
	} else {
		fmt.Printf("│ No description\n")
	}
	fmt.Printf("╰────────────────────────────────────────────────────────────╯\n")

	return nil
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := requireAPIKey(); err != nil {
		return err
	}

	p := tea.NewProgram(
		app.New(cfg),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running app: %w", err)
	}
	return nil
}
