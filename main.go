package main

import (
	"fmt"
	"os"

	"github.com/safedep/dry/log"
	"github.com/safedep/dry/obs"
	"github.com/safedep/xbom/cmd"
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

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
