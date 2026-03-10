package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/aporicho/markcli/internal/config"
	"github.com/aporicho/markcli/internal/theme"
	"github.com/aporicho/markcli/internal/tui"
)

func main() {
	root := &cobra.Command{
		Use:          "mark",
		Short:        "Mark - annotation tool for Claude Code",
		Args:         cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runOpen(cmd, args)
			}
			return cmd.Help()
		},
	}

	openCmd := &cobra.Command{
		Use:          "open <file>",
		Short:        "Open a Markdown file for annotation",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE:         runOpen,
	}

	root.AddCommand(openCmd)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runOpen(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	cfg := config.Load()

	themeNames := theme.Names()
	themeIdx := 0
	for i, n := range themeNames {
		if n == cfg.Theme {
			themeIdx = i
			break
		}
	}

	t := theme.Get(cfg.Theme)
	m := tui.New(filePath, t, themeIdx)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
