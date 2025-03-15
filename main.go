package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	kubeitCmd "github.com/komailo/kubeit/cmd/kubeit/commands"
)

func main() {
	var command *cobra.Command

	binaryName := filepath.Base(os.Args[0])

	switch binaryName {
	case "kubeit":
		command = kubeitCmd.NewCommand()
	default:
		command = kubeitCmd.NewCommand()
	}

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
