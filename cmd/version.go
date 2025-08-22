package cmd

import (
	"fmt"
	"os"

	"github.com/safedep/xbom/internal/ui"
	"github.com/safedep/xbom/internal/version"
	"github.com/safedep/xbom/pkg/common"
	"github.com/spf13/cobra"
)

var xbomTool common.ToolMetadata

func init() {

	xbomTool = common.ToolMetadata{
		Name:                 "xbom",
		Version:              version.Version,
		Purl:                 "pkg:golang/safedep/xbom@" + version.Version,
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
			fmt.Print(ui.GenerateXBOMBanner(version.Version, version.Commit))

			_, err := fmt.Fprintf(os.Stdout, "Version: %s\n", version.Version)
			if err != nil {
				return fmt.Errorf("failed to write version: %w", err)
			}

			_, err = fmt.Fprintf(os.Stdout, "CommitSHA: %s\n", version.Commit)
			if err != nil {
				return fmt.Errorf("failed to write commit SHA: %w", err)
			}

			return nil
		},
	}

	return cmd
}
