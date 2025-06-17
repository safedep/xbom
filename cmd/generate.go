package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/safedep/dry/log"
	"github.com/safedep/xbom/internal/analytics"
	"github.com/safedep/xbom/internal/command"
	"github.com/safedep/xbom/internal/ui"
	"github.com/safedep/xbom/pkg/bom"
	"github.com/safedep/xbom/pkg/codeanalysis"
	"github.com/safedep/xbom/pkg/reporter"
	"github.com/safedep/xbom/pkg/signatures"
	"github.com/spf13/cobra"
)

var (
	directory           string
	cyclonedxReportPath string
)

func NewGenerateCommand() *cobra.Command {
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

	cmd.Flags().StringVarP(&directory, "dir", "D", wd,
		"Directory for analysing and generating BOM")
	cmd.Flags().StringVarP(&cyclonedxReportPath, "bom", "", "",
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
	analytics.TrackCommandGenerate()
	command.FailOnError("generate", internalGenerate())
}

func internalGenerate() error {
	log.Infof("Generating BOM for source - %s", directory)

	// provide grouping filters using signatures.LoadSignatures("microsoft", "azure", "servicebus")
	signaturesToMatch, err := signatures.LoadAllSignatures()
	if err != nil {
		return fmt.Errorf("failed to load signatures: %w", err)
	}
	log.Debugf("Loaded %d signatures", len(signaturesToMatch))

	bomGenerator, err := bom.NewCycloneDXBomGenerator(bom.CycloneDXGeneratorConfig{
		Tool:                     xbomTool,
		Path:                     cyclonedxReportPath,
		ApplicationComponentName: path.Base(directory),
	})
	if err != nil {
		return fmt.Errorf("failed to create CycloneDX BOM generator: %w", err)
	}

	workflow := codeanalysis.NewCodeAnalysisWorkflow(codeanalysis.CodeAnalysisWorkflowConfig{
		Tool:              xbomTool,
		SourcePath:        directory,
		SignaturesToMatch: signaturesToMatch,
		Callbacks: &codeanalysis.CodeAnalysisCallbackRegistry{
			OnStart: func() error {
				ui.StartSpinner("Analysing code")
				return nil
			},
			OnFinish: func() error {
				ui.StopSpinner("✅ Code analysis completed.")
				return nil
			},
			OnErr: func(message string, err error) {
				log.Errorf("Error in code analysis workflow: %s: %v", message, err)
				ui.StopSpinner("❗Code analysis failed")
			},
		},
	})

	err = workflow.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute code analysis workflow: %w", err)
	}

	codeAnalysisFindings, err := workflow.Finish()
	if err != nil {
		return fmt.Errorf("failed to finish code analysis workflow: %w", err)
	}

	err = reporter.SummariseCodeAnalysisFindings(codeAnalysisFindings)
	if err != nil {
		return fmt.Errorf("failed to summarise code analysis findings: %w", err)
	}

	htmlPath := resolveHtmlPath()
	err = reporter.VisualiseCodeAnalysisFindings(codeAnalysisFindings, htmlPath)
	if err != nil {
		return fmt.Errorf("failed to visualise code analysis findings: %w", err)
	}

	err = bomGenerator.RecordCodeAnalysisFindings(codeAnalysisFindings)
	if err != nil {
		return fmt.Errorf("failed to record code analysis findings in BOM: %w", err)
	}

	err = bomGenerator.Finish()
	if err != nil {
		return fmt.Errorf("failed to finish BOM generation: %w", err)
	}

	fmt.Println()
	fmt.Printf("📄 BOM saved at %s\n", cyclonedxReportPath)
	fmt.Println("🔗 You can view the HTML report at:", htmlPath)

	return nil
}

func resolveHtmlPath() string {
	parts := strings.Split(cyclonedxReportPath, ".")

	if parts[len(parts)-1] == "json" {
		parts[len(parts)-1] = "html"
	} else {
		parts = append(parts, "html")
	}

	return strings.Join(parts, ".")
}
