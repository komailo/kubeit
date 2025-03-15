package commands

import (
	_ "embed"
	"fmt"
	"runtime/debug"

	"github.com/komailo/kubeit/internal/logger"
	"github.com/komailo/kubeit/internal/version"
	"github.com/spf13/cobra"
)

//go:embed assets/LICENSE
var licenseContent string

var showLicense bool

const (
	licenseFlag = "license"
)

// The dependencies that we want to print the version of
var depNames = map[string]string{
	"github.com/docker/docker": "Docker Client",
	"helm.sh/helm/v3":          "Helm",
	"k8s.io/apimachinery":      "Kubernetes API Machinery",
}

// VersionCmd provides the version of the tool
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Kubeit",
	Run: func(_ *cobra.Command, _ []string) {
		if showLicense {
			fmt.Println(licenseContent)
		}
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			logger.Warnf("Failed to read build info")
			return
		}

		fmt.Printf("Dependencies:\n")
		// Iterate over the dependencies and print their versions
		for _, dep := range bi.Deps {
			if name, ok := depNames[dep.Path]; ok {
				fmt.Printf("    %s: %s (%s)\n", name, dep.Version, dep.Sum)
			}
		}
		fmt.Printf("\n")

		version.PrintBuildInfo()

		fmt.Printf("\n")

		if !showLicense {
			fmt.Printf("For license information, run: kubeit version --%s\n", licenseFlag)
		}
	},
}

func init() {
	VersionCmd.PersistentFlags().
		BoolVar(&showLicense, licenseFlag, false, "Print the license information")
}
