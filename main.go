package main

import (
	"fmt"
	"os"

	"github.com/safedep/dry/log"
	"github.com/safedep/dry/obs"
	"github.com/safedep/xbom/cmd"
	"github.com/safedep/xbom/internal/analytics"
	"github.com/safedep/xbom/internal/ui"
	"github.com/safedep/xbom/internal/version"
	"github.com/spf13/cobra"
)

var verbose bool

func main() {
	command := &cobra.Command{
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

	command.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose logs")

	command.AddCommand(cmd.NewVersionCommand())
	command.AddCommand(cmd.NewGenerateCommand())
	command.AddCommand(cmd.NewValidateCommand())

	// Print banner on --help / -h
	command.SetHelpFunc(func(command *cobra.Command, args []string) {
		fmt.Print(ui.GenerateXBOMBanner(version.Version, version.Commit))
		fmt.Println(command.UsageString())
	})

	defer analytics.Close()

	analytics.TrackCommandRun()
	analytics.TrackCI()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
