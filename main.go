package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	reflowCmd "github.com/scorebet/reflow/cmd/reflow/commands"
)

func main() {
	var command *cobra.Command

	binaryName := filepath.Base(os.Args[0])

	switch binaryName {
	case "reflow":
		command = reflowCmd.NewCommand()
	default:
		command = reflowCmd.NewCommand()
	}

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
