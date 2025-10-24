package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/safedep/xbom/internal/ui"
	"github.com/safedep/xbom/pkg/common"
)

type SummaryReporterConfig struct {
	// MaxResults limits the number of results to display (default: 50, 0 = unlimited)
	MaxResults int
	// ShowStats toggles the statistics panel display (default: true)
	ShowStats bool
	// GroupBy specifies how to group results: "signature", "language", "file", or "" for no grouping
	GroupBy string
	// Colorize enables colored output (default: true)
	Colorize bool
}

type SummaryReporter struct {
	config          SummaryReporterConfig
	sigTable        table.Writer
	findings        *common.CodeAnalysisFindings
	totalFindings   int
	filesAffected   map[string]bool
	languageCounts  map[string]int
	signatureCounts map[string]int
}

var _ Reporter = (*SummaryReporter)(nil)

func NewSummaryReporter(config SummaryReporterConfig) (*SummaryReporter, error) {
	// Set defaults - only apply if not explicitly configured
	if config.MaxResults == 0 {
		config.MaxResults = 50
	}

	sigTable := table.NewWriter()
	sigTable.SetOutputMirror(os.Stdout)
	sigTable.SetStyle(table.StyleRounded)

	sigTable.AppendHeader(table.Row{"#", "Signature", "Language", "Condition", "Evidence File", "Location"})
	sigTable.SetTitle("🔍 Matched Signatures")

	sigTable.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:   0,
			WidthMax: 5,
		},
		{
			Number:    1,
			AutoMerge: true,
			WidthMax:  40,
		},
		{
			Number:   2,
			WidthMax: 12,
		},
		{
			Number:   3,
			WidthMax: 35,
		},
		{
			Number:   4,
			WidthMax: 30,
		},
		{
			Number:   5,
			WidthMax: 20,
		},
	})

	return &SummaryReporter{
		config:          config,
		sigTable:        sigTable,
		filesAffected:   make(map[string]bool),
		languageCounts:  make(map[string]int),
		signatureCounts: make(map[string]int),
	}, nil
}

func (r *SummaryReporter) Name() string {
	return "summary"
}

// colorize applies color to text if colorization is enabled
func (r *SummaryReporter) colorize(colorFunc func(a ...interface{}) string, text string) string {
	if r.config.Colorize {
		return colorFunc(text)
	}
	return text
}

// getLanguageColor returns a color for a given language
func (r *SummaryReporter) getLanguageColor(language string) func(a ...interface{}) string {
	languageColors := map[string]func(a ...interface{}) string{
		"python":     color.New(color.FgYellow).SprintFunc(),
		"javascript": color.New(color.FgYellow, color.Bold).SprintFunc(),
		"typescript": color.New(color.FgBlue, color.Bold).SprintFunc(),
		"java":       color.New(color.FgRed, color.Bold).SprintFunc(),
		"go":         color.New(color.FgCyan, color.Bold).SprintFunc(),
		"rust":       color.New(color.FgRed).SprintFunc(),
		"ruby":       color.New(color.FgRed).SprintFunc(),
	}

	langLower := strings.ToLower(language)
	if colorFunc, ok := languageColors[langLower]; ok {
		return colorFunc
	}

	return color.New(color.FgWhite).SprintFunc()
}

func (r *SummaryReporter) RecordCodeAnalysisFindings(codeAnalysisFindings *common.CodeAnalysisFindings) error {
	r.findings = codeAnalysisFindings

	cyan := color.New(color.FgCyan).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	rowNum := 0
	for _, signatureResults := range codeAnalysisFindings.SignatureWiseMatchResults {
		for _, signatureMatchResult := range signatureResults {
			// Collect statistics
			r.filesAffected[signatureMatchResult.FilePath] = true
			r.languageCounts[string(signatureMatchResult.MatchedLanguageCode)]++
			r.signatureCounts[signatureMatchResult.MatchedSignature.Id]++

			for _, condition := range signatureMatchResult.MatchedConditions {
				for _, evidence := range condition.Evidences {
					r.totalFindings++

					// Check if we've reached the limit
					if r.config.MaxResults > 0 && rowNum >= r.config.MaxResults {
						continue
					}

					rowNum++

					evidenceDetailString := "Unknown"
					evidenceMetadata := evidence.Metadata(signatureMatchResult.TreeData)
					if evidenceMetadata.CallerIdentifierMetadata != nil {
						evidenceDetailString = fmt.Sprintf(
							"L%d:%d-L%d:%d",
							evidenceMetadata.CallerIdentifierMetadata.StartLine+1,
							evidenceMetadata.CallerIdentifierMetadata.StartColumn+1,
							evidenceMetadata.CallerIdentifierMetadata.EndLine+1,
							evidenceMetadata.CallerIdentifierMetadata.EndColumn+1,
						)
					}

					// Format signature ID with color
					sigId := r.colorize(cyan, signatureMatchResult.MatchedSignature.Id)

					// Format language with appropriate color
					langColor := r.getLanguageColor(string(signatureMatchResult.MatchedLanguageCode))
					lang := r.colorize(langColor, string(signatureMatchResult.MatchedLanguageCode))

					// Format condition
					conditionStr := fmt.Sprintf("%s:\n%s", condition.Condition.Type, condition.Condition.Value)

					// Format file path (base name for readability)
					fileName := filepath.Base(signatureMatchResult.FilePath)
					filePath := r.colorize(dim, fileName)

					// Format location with color
					location := r.colorize(yellow, evidenceDetailString)

					r.sigTable.AppendRow(table.Row{
						rowNum,
						sigId,
						lang,
						conditionStr,
						filePath,
						location,
					})
					r.sigTable.AppendSeparator()
				}
			}
		}
	}

	return nil
}

func (r *SummaryReporter) Finish() error {
	ui.Println()

	// Display statistics if enabled
	if r.config.ShowStats && r.totalFindings > 0 {
		r.renderStatistics()
		ui.Println()
	}

	// Render table if there are findings
	if r.totalFindings > 0 {
		r.sigTable.Render()

		// Show truncation message if needed
		if r.config.MaxResults > 0 && r.totalFindings > r.config.MaxResults {
			ui.Println()
			truncated := r.totalFindings - r.config.MaxResults
			gray := color.New(color.Faint).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()
			msg := fmt.Sprintf("%s %s",
				r.colorize(gray, fmt.Sprintf("... and %d more findings.", truncated)),
				r.colorize(yellow, "Use --html flag to view all results."))
			ui.Println(msg)
		}
	} else {
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		ui.Println(r.colorize(green, "✓ No signature matches found!"))
	}

	return nil
}

func (r *SummaryReporter) renderStatistics() {
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Create statistics table
	statsTable := table.NewWriter()
	statsTable.SetOutputMirror(os.Stdout)
	statsTable.SetStyle(table.StyleRounded)
	statsTable.SetTitle(r.colorize(bold, "📊 Analysis Summary"))

	statsTable.AppendRow(table.Row{
		r.colorize(cyan, "Total Findings:"),
		r.colorize(bold, fmt.Sprintf("%d", r.totalFindings)),
	})

	statsTable.AppendRow(table.Row{
		r.colorize(cyan, "Unique Signatures:"),
		r.colorize(bold, fmt.Sprintf("%d", len(r.signatureCounts))),
	})

	statsTable.AppendRow(table.Row{
		r.colorize(cyan, "Files Affected:"),
		r.colorize(bold, fmt.Sprintf("%d", len(r.filesAffected))),
	})

	statsTable.AppendRow(table.Row{
		r.colorize(cyan, "Languages Detected:"),
		r.colorize(bold, fmt.Sprintf("%d", len(r.languageCounts))),
	})

	statsTable.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:   0,
			WidthMax: 35,
			Align:    text.AlignRight,
		},
		{
			Number:   1,
			WidthMax: 25,
			Align:    text.AlignLeft,
		},
	})

	statsTable.Render()

	// Show top signatures
	if len(r.signatureCounts) > 0 {
		ui.Println()
		ui.Println(r.colorize(yellow, "🏆 Top Matched Signatures:"))

		// Sort signatures by count
		type sigCount struct {
			id    string
			count int
		}
		var sigs []sigCount
		for id, count := range r.signatureCounts {
			sigs = append(sigs, sigCount{id, count})
		}
		sort.Slice(sigs, func(i, j int) bool {
			return sigs[i].count > sigs[j].count
		})

		// Show top 5
		limit := 5
		if len(sigs) < limit {
			limit = len(sigs)
		}
		for i := 0; i < limit; i++ {
			ui.Println(fmt.Sprintf("  %s %s %s",
				r.colorize(green, fmt.Sprintf("%d.", i+1)),
				r.colorize(cyan, sigs[i].id),
				r.colorize(color.New(color.Faint).SprintFunc(), fmt.Sprintf("(%d matches)", sigs[i].count))))
		}
	}

	// Show language breakdown
	if len(r.languageCounts) > 0 {
		ui.Println()
		ui.Println(r.colorize(yellow, "💬 Languages:"))

		var langs []string
		for lang := range r.languageCounts {
			langs = append(langs, lang)
		}
		sort.Strings(langs)

		for _, lang := range langs {
			langColor := r.getLanguageColor(lang)
			ui.Println(fmt.Sprintf("  • %s %s",
				r.colorize(langColor, lang),
				r.colorize(color.New(color.Faint).SprintFunc(), fmt.Sprintf("(%d findings)", r.languageCounts[lang]))))
		}
	}
}
