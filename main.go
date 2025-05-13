package main

import (
	"fmt"
	"os"

	"github.com/safedep/dry/log"
	"github.com/safedep/dry/obs"
	"github.com/safedep/xbom/pkg/common"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	debug   bool
	logFile string
)

var xbomTool = common.ToolMetadata{
	Name:                 "xbom",
	Version:              version,
	Purl:                 "pkg:golang/safedep/xbom@" + version,
	InformationURI:       "https://github.com/safedep/xbom",
	VendorName:           "Safedep",
	VendorInformationURI: "https://safedep.io",
}

func main() {
	cmd := &cobra.Command{
		Use:              "xbom [OPTIONS] COMMAND [ARG...]",
		Short:            "[ Generate BOMs enriched with AI, ML, SaaS, Cloud and more ]",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			return fmt.Errorf("xbom: %s is not a valid command", args[0])
		},
	}

	cobra.OnInitialize(func() {
		log.InitCliLogger(obs.AppServiceName("xbom"), obs.AppServiceEnv("dev"))
	})

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose logs")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Show debug logs")
	cmd.PersistentFlags().StringVarP(&logFile, "log", "l", "", "Write command logs to file, use '-' for stdout")

	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newGenerateCommand())
	cmd.AddCommand(newValidateCommand())

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
