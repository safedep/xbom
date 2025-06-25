package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/safedep/dry/log"
	"github.com/safedep/xbom/internal/analytics"
	"github.com/safedep/xbom/internal/command"
	"github.com/safedep/xbom/internal/ui"
	"github.com/safedep/xbom/pkg/codeanalysis"
	"github.com/safedep/xbom/pkg/reporter"
	"github.com/safedep/xbom/pkg/signatures"
	"github.com/spf13/cobra"
)

var (
	codeDirectory       string
	cyclonedxReportPath string
	htmlReportPath      string
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

	cmd.Flags().StringVarP(&codeDirectory, "dir", "D", wd,
		"Directory for analysing and generating BOM")
	cmd.Flags().StringVarP(&cyclonedxReportPath, "bom", "", "",
		"Generate CycloneDX BOM to file")
	cmd.Flags().StringVarP(&htmlReportPath, "html", "", "",
		"Generate HTML report to file")

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
	log.Infof("Generating BOM for source - %s", codeDirectory)

	// provide grouping filters using signatures.LoadSignatures("microsoft", "azure", "servicebus")
	signaturesToMatch, err := signatures.LoadAllSignatures()
	if err != nil {
		return fmt.Errorf("failed to load signatures: %w", err)
	}
	log.Debugf("Loaded %d signatures", len(signaturesToMatch))

	reporters := []reporter.Reporter{}

	summaryReporter, err := reporter.NewSummaryReporter(reporter.SummaryReporterConfig{})
	if err != nil {
		return fmt.Errorf("failed to create summary reporter: %w", err)
	}
	reporters = append(reporters, summaryReporter)

	if cyclonedxReportPath != "" {
		cdxReporter, err := reporter.NewCycloneDXBomReporter(reporter.CycloneDXReporterConfig{
			Tool:                     xbomTool,
			Path:                     cyclonedxReportPath,
			ApplicationComponentName: path.Base(codeDirectory),
		})
		if err != nil {
			return fmt.Errorf("failed to create CycloneDX reporter: %w", err)
		}
		reporters = append(reporters, cdxReporter)
	}

	if htmlReportPath != "" {
		htmlReporter, err := reporter.NewHTMLReporter(reporter.HTMLReporterConfig{
			HtmlReportPath: htmlReportPath,
		})
		if err != nil {
			return fmt.Errorf("failed to create HTML reporter: %w", err)
		}
		reporters = append(reporters, htmlReporter)
	}

	workflow := codeanalysis.NewCodeAnalysisWorkflow(
		codeanalysis.CodeAnalysisWorkflowConfig{
			Tool:              xbomTool,
			SourcePath:        codeDirectory,
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
		},
		reporters,
	)

	// If xbom is used as a library, we may use the finalised findings here
	_, err = workflow.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute code analysis workflow: %w", err)
	}

	// Nudge user to visualise the results
	if htmlReportPath == "" {
		fmt.Println("\nTip: You can visualise the report as HTML using \"--html\" flag.")
		fmt.Println("Example: xbom generate --html /tmp/report.html")
	}

	return nil
}
