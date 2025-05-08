package main

import (
	"fmt"
	"os"
	"path"

	"github.com/safedep/dry/log"
	"github.com/safedep/xbom/internal/command"
	"github.com/safedep/xbom/internal/ui"
	"github.com/safedep/xbom/pkg/bom"
	"github.com/safedep/xbom/pkg/codeanalysis"
	"github.com/safedep/xbom/pkg/signatures"
	"github.com/spf13/cobra"
)

var (
	codePath            string
	cyclonedxReportPath string
)

func newGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate BOMs enriched with AI, ML, SaaS, Cloud and more",
		RunE: func(cmd *cobra.Command, args []string) error {
			generate()
			return nil
		},
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cmd.Flags().StringVarP(&codePath, "code", "C", wd,
		"Source code for generating BOM")
	cmd.Flags().StringVarP(&cyclonedxReportPath, "cdx", "", "",
		"Generate CycloneDX BOM to file")

	_ = cmd.MarkFlagRequired("cdx")

	// Add validations that should trigger a fail fast condition
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		err := func() error {
			return nil
		}()

		command.FailOnError("pre-scan", err)
	}

	return cmd
}

func generate() {
	command.FailOnError("generate", internalGenerate())
}

func internalGenerate() error {
	log.Infof("Generating BOM for source - %s", codePath)

	// provide grouping filters using signatures.LoadSignatures("microsoft", "azure", "servicebus")
	signaturesToMatch, err := signatures.LoadAllSignatures()
	if err != nil {
		return fmt.Errorf("failed to load signatures: %w", err)
	}
	log.Debugf("Loaded %d signatures", len(signaturesToMatch))

	bomGenerator, err := bom.NewCycloneDXBomGenerator(bom.CycloneDXGeneratorConfig{
		Tool:                     xbomTool,
		Path:                     cyclonedxReportPath,
		ApplicationComponentName: path.Base(codePath),
	})
	if err != nil {
		return fmt.Errorf("failed to create CycloneDX BOM generator: %w", err)
	}

	workflow := codeanalysis.NewCodeAnalysisWorkflow(codeanalysis.CodeAnalysisWorkflowConfig{
		Tool:              xbomTool,
		SourcePath:        codePath,
		SignaturesToMatch: signaturesToMatch,
		Callbacks: &codeanalysis.CodeAnalysisCallbackRegistry{
			OnStart: func() error {
				ui.StartSpinner("Analysing code")
				return nil
			},
			OnFinish: func() error {
				ui.StopSpinner("✅ Code analysis completed")
				return nil
			},
			OnErr: func(message string, err error) error {
				log.Errorf("Error in code analysis workflow: %s: %v", message, err)
				ui.StopSpinner("❗Code analysis failed")
				return nil
			},
		},
	})

	redirectLogToFile(logFile)

	workflow.Execute()
	codeAnalysisFindings := workflow.Finish(true)

	bomGenerator.RecordCodeAnalysisFindings(codeAnalysisFindings)
	bomGenerator.Finish()

	return nil
}
