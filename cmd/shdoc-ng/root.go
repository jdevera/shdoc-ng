package main

import (
	"os"

	"github.com/spf13/cobra"
)

// version is set at build time via ldflags.
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "shdoc-ng",
	Version: version,
	Short:   "Generate documentation from annotated shell scripts",
	Long: `shdoc-ng reads shell scripts and produces documentation by extracting
structured comment blocks written above shell functions. Supports tags like
@description, @arg, @option, @exitcode, @example, and more.

When invoked without a subcommand and stdin is a pipe, it behaves as
"shdoc-ng generate" for convenient use in pipelines:

  shdoc-ng < script.sh > docs.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If stdin is a pipe, behave as "generate".
		fi, err := os.Stdin.Stat()
		if err != nil {
			return cmd.Help()
		}
		if fi.Mode()&os.ModeCharDevice == 0 {
			return generateCmd.RunE(generateCmd, args)
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion script",
	Long: `Generate a completion script for the specified shell.

  # bash - load in current session
  source <(shdoc-ng completion bash)

  # bash - install permanently
  shdoc-ng completion bash > /etc/bash_completion.d/shdoc-ng

  # zsh - install permanently (then restart your shell)
  shdoc-ng completion zsh > "${fpath[1]}/_shdoc-ng"
  # If that directory doesn't work, try:
  #   mkdir -p ~/.zsh/completions
  #   shdoc-ng completion zsh > ~/.zsh/completions/_shdoc-ng
  #   # then add to .zshrc: fpath=(~/.zsh/completions $fpath); autoload -Uz compinit && compinit

  # fish - load in current session
  shdoc-ng completion fish | source

  # fish - install permanently
  shdoc-ng completion fish > ~/.config/fish/completions/shdoc-ng.fish`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		default:
			return cmd.Help()
		}
	},
}
