package cmd

import (
	"context"
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
	packageURL          string
	appName             string
	codeDirectory       string
	cyclonedxReportPath string
	htmlReportPath      string
	markdownReportPath  string
	summaryMaxResults   int
	summaryNoStats      bool
	summaryNoColor      bool
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
	cmd.Flags().StringVarP(&packageURL, "purl", "P", "",
		"Package URL of a supported OSS package (eg. pkg:/npm/express@4.17.1")
	cmd.Flags().StringVarP(&appName, "app-name", "", "",
		"App name to include in CycloneDX BOM")
	cmd.Flags().StringVarP(&cyclonedxReportPath, "bom", "", "",
		"Generate CycloneDX BOM to file")
	cmd.Flags().StringVarP(&htmlReportPath, "report-html", "", "",
		"Generate HTML report to file")
	cmd.Flags().StringVarP(&markdownReportPath, "report-markdown", "", "",
		"Generate Markdown report to file")
	cmd.Flags().IntVarP(&summaryMaxResults, "summary-limit", "", 20,
		"Maximum number of results to display in summary (0 for unlimited)")
	cmd.Flags().BoolVarP(&summaryNoStats, "summary-no-stats", "", false,
		"Disable statistics panel in summary output")
	cmd.Flags().BoolVarP(&summaryNoColor, "summary-no-color", "", false,
		"Disable colored output in summary")

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
	command.FailOnError("generate", internalGenerateMulti())
}

// internalGenerateMulti handles multiple input adapters before invoking the
// core scanning workflow
func internalGenerateMulti() error {
	// Start with different supported adapters based on args
	if packageURL != "" {
		return internalGeneratePurl()
	}

	// Fallback to the last option ie. local directory
	if appName == "" {
		appName = path.Base(codeDirectory)
	}

	return internalGenerateDirectory(appName, codeDirectory)
}

// internalGeneratePurl setup a local cache for a package
// identified by its PURL for scanning. It also cleanup the local
// cache after the scanning process.
func internalGeneratePurl() error {
	pullResponse, err := command.PackagePull(context.Background(), command.PackagePullRequest{
		PURL: packageURL,
	})
	if err != nil {
		return fmt.Errorf("failed to pull package: %w", err)
	}

	defer func() {
		if err := pullResponse.Close(); err != nil {
			log.Errorf("failed to cleanup package: %v", err)
		}
	}()

	localPath, err := pullResponse.LocalPath()
	if err != nil {
		return fmt.Errorf("failed to find local path for package: %w", err)
	}

	if appName == "" {
		appName = packageURL
	}

	return internalGenerateDirectory(appName, localPath)
}

// internalGenerate executes the core scanning workflow to generate an XBOM report
func internalGenerateDirectory(appName, codeDir string) error {
	log.Infof("Generating BOM for source - %s", codeDir)

	// provide grouping filters using signatures.LoadSignatures("microsoft", "azure", "servicebus")
	signaturesToMatch, err := signatures.LoadAllSignatures()
	if err != nil {
		return fmt.Errorf("failed to load signatures: %w", err)
	}

	log.Debugf("Loaded %d signatures", len(signaturesToMatch))

	reporters := []reporter.Reporter{}

	summaryReporter, err := reporter.NewSummaryReporter(reporter.SummaryReporterConfig{
		MaxResults: summaryMaxResults,
		ShowStats:  !summaryNoStats,
		Colorize:   !summaryNoColor,
	})
	if err != nil {
		return fmt.Errorf("failed to create summary reporter: %w", err)
	}
	reporters = append(reporters, summaryReporter)

	if cyclonedxReportPath != "" {
		cdxReporter, err := reporter.NewCycloneDXBomReporter(reporter.CycloneDXReporterConfig{
			Tool:                     xbomTool,
			Path:                     cyclonedxReportPath,
			ApplicationComponentName: appName,
		})
		if err != nil {
			return fmt.Errorf("failed to create CycloneDX reporter: %w", err)
		}
		reporters = append(reporters, cdxReporter)
	}

	if htmlReportPath != "" {
		htmlReporter, err := reporter.NewHTMLReporter(reporter.HTMLReporterConfig{
			HTMLReportPath: htmlReportPath,
		})
		if err != nil {
			return fmt.Errorf("failed to create HTML reporter: %w", err)
		}
		reporters = append(reporters, htmlReporter)
	}

	if markdownReportPath != "" {
		markdownReporter, err := reporter.NewMarkdownReporter(reporter.MarkdownReporterConfig{
			OutputPath: markdownReportPath,
		})
		if err != nil {
			return fmt.Errorf("failed to create Markdown reporter: %w", err)
		}
		reporters = append(reporters, markdownReporter)
	}

	workflow := codeanalysis.NewCodeAnalysisWorkflow(
		codeanalysis.CodeAnalysisWorkflowConfig{
			Tool:              xbomTool,
			SourcePath:        codeDir,
			SignaturesToMatch: signaturesToMatch,
			Callbacks: codeanalysis.CodeAnalysisCallbackRegistry{
				OnStart: func() error {
					ui.StartSpinner("Analyzing code")
					return nil
				},
				OnFinish: func() error {
					ui.StopSpinner("✅ Code analysis completed.")
					return nil
				},
				OnErr: func(message string, err error) {
					log.Errorf("Error in code analysis workflow: %s: %v", message, err)
					ui.StopSpinner(fmt.Sprintf("❗Code analysis failed with error: %s", err.Error()))
				},
			},
		},
		reporters,
	)

	// If xbom is used as a library, we may use the finalized findings here
	_, err = workflow.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute code analysis workflow: %w", err)
	}

	// Nudge user to visualise the results
	if htmlReportPath == "" && markdownReportPath == "" {
		ui.Println()
		ui.Println("Tip: You can save the report to a file using \"--report-html\" or \"--report-markdown\" flags.")
		ui.Println("Examples:")
		ui.Println("  xbom generate --report-html /tmp/report.html")
		ui.Println("  xbom generate --report-markdown /tmp/report.md")
	}

	return nil
}
