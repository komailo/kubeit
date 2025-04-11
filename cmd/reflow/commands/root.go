package commands

import (
	"fmt"

	"github.com/scorebet/reflow/common"
	"github.com/scorebet/reflow/internal/logger"

	"github.com/spf13/cobra"
)

// Global options accessible by all subcommands
var globalSetOpts globalOptions

// RootCmd is the base command
var RootCmd = &cobra.Command{
	Use:   common.MainCLIName,
	Short: "CLI tool to generate and manage Kubernetes deployment configurations",
	Long: fmt.Sprintf(`%s is a CLI tool designed for service teams to simplify
the generation and management of Kubernetes deployment configurations.

It allows teams to define infrastructure in a minimal YAML format
and transforms it into fully rendered Kubernetes objects.

Use 'reflow generate' to convert a %s configuration into
Kubernetes manifests for deployment.`, common.AppName, common.AppName),
}

func init() {
	// Global verbosity flag
	RootCmd.PersistentFlags().CountVarP(
		&globalSetOpts.Verbosity,
		"verbose",
		"v",
		"Increase verbosity (-v = debug, -vv = trace). By default only info, warnings and errors are shown.",
	)

	// Initialize the logger after flags are parsed
	cobra.OnInitialize(initLogger)

	// Register subcommands
	RootCmd.AddCommand(GenerateCmd)
	RootCmd.AddCommand(VersionCmd)
	RootCmd.SilenceUsage = true
}

func initLogger() {
	logger.SetLevelFromVerbosity(globalSetOpts.Verbosity + 1)
}

// NewRootCommand returns the root command instead of executing it
func NewCommand() *cobra.Command {
	return RootCmd
}
