package commands

import (
	"fmt"
	"runtime/debug"

	_ "embed" // Ensure the embed package is imported

	"github.com/komailo/kubeit/common"
	"github.com/komailo/kubeit/internal/logger"
	"github.com/spf13/cobra"
)

//go:embed assets/LICENSE
var licenseContent string

var showLicense bool

const (
	licenseFlag = "license"
)

var depNames = map[string]string{
	"github.com/docker/docker": "Docker Client",
	"helm.sh/helm/v3":          "Helm",
	"k8s.io/client-go":         "Kubernetes API machinery",
}

// VersionCmd provides the version of the tool
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Kubeit",
	Run: func(cmd *cobra.Command, args []string) {
		if showLicense {
			fmt.Println(licenseContent)
		}
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			logger.Warnf("Failed to read build info")
			return
		}

		// Iterate over the dependencies and print their versions
		for _, dep := range bi.Deps {
			if name, ok := depNames[dep.Path]; ok {
				fmt.Printf("%s Version: %s\n", name, dep.Version)
			}
		}
		fmt.Printf("Kubeit version: %s\n\n", common.Version)

		if !showLicense {
			fmt.Printf("For license information, run: kubeit version --%s\n", licenseFlag)
		}

	},
}

func init() {
	VersionCmd.PersistentFlags().BoolVar(&showLicense, licenseFlag, false, "Print the license information")

}
