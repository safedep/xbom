package cmd

import (
	"fmt"

	"github.com/safedep/xbom/internal/command"
	"github.com/safedep/xbom/pkg/signatures"
	"github.com/spf13/cobra"
)

func NewValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate signatures",
		RunE: func(cmd *cobra.Command, args []string) error {
			validate()
			return nil
		},
	}

	// Add validations that should trigger a fail fast condition
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		err := func() error {
			return nil
		}()

		command.FailOnError("pre-scan", err)
	}

	return cmd
}

func validate() {
	command.FailOnError("validate", internalValidate())
}

func internalValidate() error {
	_, err := signatures.LoadAllSignatures()
	if err == nil {
		fmt.Println("✅ Signatures valid")
	} else {
		fmt.Println("❌ Signatures invalid")
	}
	return err
}
