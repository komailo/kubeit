package main

import (
	"os"
	"path/filepath"

	kubeitCmd "github.com/komailo/kubeit/cmd/kubeit/commands"
	"github.com/spf13/cobra"
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
