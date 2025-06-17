package cmd

import (
	"fmt"
	"os"
	runtimeDebug "runtime/debug"

	"github.com/safedep/xbom/pkg/common"
	"github.com/spf13/cobra"
)

// When building with CI or Make, version is set using `ldflags`
var (
	version string
	commit  string
)

var xbomTool common.ToolMetadata

func init() {
	// Only use buildInfo if version wasn't set by ldflags, that is its being build by `go install`
	if version == "" {
		// Main.Version is based on the version control system tag or commit.
		// This useful when app is build with `go install`
		// See: https://antonz.org/go-1-24/#main-modules-version
		buildInfo, _ := runtimeDebug.ReadBuildInfo()
		version = buildInfo.Main.Version
	}

	xbomTool = common.ToolMetadata{
		Name:                 "xbom",
		Version:              version,
		Purl:                 "pkg:golang/safedep/xbom@" + version,
		InformationURI:       "https://github.com/safedep/xbom",
		VendorName:           "SafeDep",
		VendorInformationURI: "https://safedep.io",
	}
}

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version and build information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(os.Stdout, "Version: %s\n", version)
			if err != nil {
				return fmt.Errorf("failed to write version: %w", err)
			}

			_, err = fmt.Fprintf(os.Stdout, "CommitSHA: %s\n", commit)
			if err != nil {
				return fmt.Errorf("failed to write commit SHA: %w", err)
			}

			return nil
		},
	}

	return cmd
}
