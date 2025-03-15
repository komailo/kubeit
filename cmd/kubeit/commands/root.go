package commands

import (
	"github.com/komailo/kubeit/common"
	"github.com/komailo/kubeit/internal/logger"

	"github.com/spf13/cobra"
)

// Global options accessible by all subcommands
var globalSetOpts globalOptions

// RootCmd is the base command
var RootCmd = &cobra.Command{
	Use:   common.KubeitCLIName,
	Short: "CLI tool to generate and manage Kubernetes deployment configurations",
	Long: `Kubeit is a CLI tool designed for service teams to simplify 
the generation and management of Kubernetes deployment configurations. 

It allows teams to define infrastructure in a minimal YAML format 
and transforms it into fully rendered Kubernetes objects.

Use 'kubeit generate' to convert a KubeIt configuration into 
Kubernetes manifests for deployment.`,
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
