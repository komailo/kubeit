package version

import (
	"fmt"
	"reflect"
	"runtime"
)

var (
	// provided by .Date in goreleaser
	buildDate = ""

	// provided by .FullCommit in goreleaser
	gitCommit = ""

	// gitSummary is the output of git describe --tags --dirty --always
	// This is provided by .Summary in goreleaser
	gitSummary = ""

	// gitTreeState is the state of the git tree, clean or dirty
	// provided by .GitTreeState in goreleaser
	gitTreeState = ""

	// provided by .Version in goreleaser
	version = "v0.0"
)

// BuildInfo describes the compile time build information.
type BuildInfo struct {
	BuildDate    string
	GitCommit    string
	GitSummary   string
	GitTreeState string
	GoVersion    string
	Version      string
}

// Get returns build info that was set at compile time.
func GetBuildInfo() BuildInfo {
	buildInfo := BuildInfo{
		BuildDate:    buildDate,
		GitCommit:    gitCommit,
		GitSummary:   gitSummary,
		GitTreeState: gitTreeState,
		GoVersion:    runtime.Version(),
		Version:      version,
	}

	return buildInfo
}

// PrintBuildInfo prints the build information.
// Its a handy function that version commands can use to print the build information.
func PrintBuildInfo() {
	buildInfo := GetBuildInfo()

	fmt.Printf("Build Info:\n")

	v := reflect.ValueOf(buildInfo)
	typeOfBuildInfo := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fmt.Printf("    %s: %v\n", typeOfBuildInfo.Field(i).Name, field.Interface())
	}
}
