package main

import (
	"os"
	"path/filepath"

	"github.com/komailo/kubeit/cmd/kubeit/commands"
	"github.com/spf13/cobra"
)

func main() {
	var command *cobra.Command

	binaryName := filepath.Base(os.Args[0])

	switch binaryName {
	case "kubeit":
		command = commands.NewCommand()
	default:
		command = commands.NewCommand()
	}

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
