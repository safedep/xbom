package reporter

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/safedep/dry/log"
	"github.com/safedep/xbom/pkg/common"
)

type MarkdownReporterConfig struct {
	OutputPath          string // Path to save the markdown report
	SnippetBeforeLines  int    // Number of context lines to show before match (default: 3)
	SnippetAfterLines   int    // Number of context lines to show after match (default: 3)
	SnippetMaxBytes     int    // Max total bytes for snippet (default: 5120 = 5KB)
	SnippetMaxLineChars int    // Max characters per line (default: 500)

	// Boolean flags for section control (all true by default)
	ShowExecutiveSummary  bool // Show executive summary section
	ShowStatistics        bool // Show statistics section
	ShowTopSignatures     bool // Show top matched signatures section
	ShowLanguageBreakdown bool // Show language breakdown section
	ShowDetailedFindings  bool // Show detailed findings section
}

type MarkdownReporter struct {
	config     MarkdownReporterConfig
	findings   *common.CodeAnalysisFindings
	statistics *reportStatistics
}

type reportStatistics struct {
	totalFindings   int
	filesAffected   map[string]bool
	languageCounts  map[string]int
	signatureCounts map[string]int
}

type signatureDetail struct {
	ID                string
	Description       string
	Tags              []string
	TotalMatches      int
	FileOccurrences   []fileOccurrence
}

type fileOccurrence struct {
	FilePath string
	Language string
	Matches  []matchDetail
}

type matchDetail struct {
	Condition string
	Snippet   *snippetInfo
}

type snippetInfo struct {
	Lines             []snippetLineData
	RawContent        string
	WasTruncated      bool
	SourceUnavailable bool
}

var _ Reporter = (*MarkdownReporter)(nil)

//go:embed templates/report.md
var markdownTemplateFS embed.FS

func NewMarkdownReporter(config MarkdownReporterConfig) (*MarkdownReporter, error) {
	// Set defaults for snippet configuration
	if config.SnippetBeforeLines == 0 {
		config.SnippetBeforeLines = 3
	}
	if config.SnippetAfterLines == 0 {
		config.SnippetAfterLines = 3
	}
	if config.SnippetMaxBytes == 0 {
		config.SnippetMaxBytes = 5120 // 5KB
	}
	if config.SnippetMaxLineChars == 0 {
		config.SnippetMaxLineChars = 500
	}

	// Set defaults for section visibility - default all to enabled
	// If at least one section flag is explicitly true, assume user is configuring sections
	// Otherwise, if all are false (zero values), enable all sections by default
	hasAnySectionEnabled := config.ShowExecutiveSummary || config.ShowStatistics ||
		config.ShowTopSignatures || config.ShowLanguageBreakdown ||
		config.ShowDetailedFindings

	if !hasAnySectionEnabled {
		// All are false (likely zero values) - enable all sections by default
		config.ShowExecutiveSummary = true
		config.ShowStatistics = true
		config.ShowTopSignatures = true
		config.ShowLanguageBreakdown = true
		config.ShowDetailedFindings = true
	}
	// If any section is enabled, respect the user's configuration as-is

	return &MarkdownReporter{
		config: config,
		statistics: &reportStatistics{
			filesAffected:   make(map[string]bool),
			languageCounts:  make(map[string]int),
			signatureCounts: make(map[string]int),
		},
	}, nil
}

func (r *MarkdownReporter) Name() string {
	return "markdown"
}

func (r *MarkdownReporter) RecordCodeAnalysisFindings(codeAnalysisFindings *common.CodeAnalysisFindings) error {
	r.findings = codeAnalysisFindings

	// Calculate statistics
	for _, signatureResults := range codeAnalysisFindings.SignatureWiseMatchResults {
		for _, signatureMatchResult := range signatureResults {
			r.statistics.filesAffected[signatureMatchResult.FilePath] = true
			r.statistics.languageCounts[string(signatureMatchResult.MatchedLanguageCode)]++

			for _, condition := range signatureMatchResult.MatchedConditions {
				for range condition.Evidences {
					r.statistics.totalFindings++
					r.statistics.signatureCounts[signatureMatchResult.MatchedSignature.Id]++
				}
			}
		}
	}

	return nil
}

func (r *MarkdownReporter) Finish() error {
	reportData := r.prepareReportData()

	// Load and parse template
	templateContent, err := markdownTemplateFS.ReadFile("templates/report.md")
	if err != nil {
		return fmt.Errorf("failed to load markdown template: %w", err)
	}

	tmpl, err := template.New("report").Funcs(r.getTemplateFuncs()).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse markdown template: %w", err)
	}

	// Create output file
	f, err := os.Create(r.config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create markdown file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("failed to close markdown report file: %v", err)
		}
	}()

	// Execute template
	if err := tmpl.Execute(f, reportData); err != nil {
		return fmt.Errorf("failed to execute markdown template: %w", err)
	}

	fmt.Println("Markdown report generated at:", r.config.OutputPath)
	return nil
}

func (r *MarkdownReporter) prepareReportData() map[string]interface{} {
	return map[string]interface{}{
		"GeneratedAt":           time.Now().Format(time.RFC3339),
		"Config":                r.config,
		"Statistics":            r.prepareStatistics(),
		"TopSignatures":         r.prepareTopSignatures(),
		"LanguageBreakdown":     r.prepareLanguageBreakdown(),
		"DetailedFindings":      r.prepareDetailedFindings(),
		"HasFindings":           r.statistics.totalFindings > 0,
	}
}

func (r *MarkdownReporter) prepareStatistics() map[string]interface{} {
	return map[string]interface{}{
		"TotalFindings":    r.statistics.totalFindings,
		"UniqueSignatures": len(r.statistics.signatureCounts),
		"FilesAffected":    len(r.statistics.filesAffected),
		"LanguagesDetected": len(r.statistics.languageCounts),
	}
}

func (r *MarkdownReporter) prepareTopSignatures() []map[string]interface{} {
	type sigCount struct {
		id    string
		count int
	}

	var sigs []sigCount
	for id, count := range r.statistics.signatureCounts {
		sigs = append(sigs, sigCount{id, count})
	}

	// Sort by count descending
	sort.Slice(sigs, func(i, j int) bool {
		return sigs[i].count > sigs[j].count
	})

	// Return top 10
	limit := 10
	if len(sigs) < limit {
		limit = len(sigs)
	}

	result := make([]map[string]interface{}, limit)
	for i := 0; i < limit; i++ {
		result[i] = map[string]interface{}{
			"Rank":  i + 1,
			"ID":    sigs[i].id,
			"Count": sigs[i].count,
		}
	}

	return result
}

func (r *MarkdownReporter) prepareLanguageBreakdown() []map[string]interface{} {
	var langs []string
	for lang := range r.statistics.languageCounts {
		langs = append(langs, lang)
	}
	sort.Strings(langs)

	result := make([]map[string]interface{}, len(langs))
	for i, lang := range langs {
		result[i] = map[string]interface{}{
			"Language": lang,
			"Count":    r.statistics.languageCounts[lang],
		}
	}

	return result
}

func (r *MarkdownReporter) prepareDetailedFindings() []signatureDetail {
	if r.findings == nil {
		return []signatureDetail{}
	}

	sigMap := make(map[string]*signatureDetail)

	for _, signatureResults := range r.findings.SignatureWiseMatchResults {
		for _, signatureMatchResult := range signatureResults {
			sig := signatureMatchResult.MatchedSignature
			sigID := sig.Id

			// Initialize signature detail if not exists
			if _, ok := sigMap[sigID]; !ok {
				sigMap[sigID] = &signatureDetail{
					ID:              sigID,
					Description:     sig.Description,
					Tags:            sig.Tags,
					FileOccurrences: []fileOccurrence{},
				}
			}

			// Create file occurrence
			fileOcc := fileOccurrence{
				FilePath: signatureMatchResult.FilePath,
				Language: string(signatureMatchResult.MatchedLanguageCode),
				Matches:  []matchDetail{},
			}

			for _, condition := range signatureMatchResult.MatchedConditions {
				for _, evidence := range condition.Evidences {
					evidenceMetadata := evidence.Metadata(signatureMatchResult.TreeData)

					conditionStr := fmt.Sprintf("%s: %s",
						condition.Condition.Type,
						strings.ReplaceAll(condition.Condition.Value, "\n", " "))

					match := matchDetail{
						Condition: conditionStr,
					}

					// Extract snippet if available
					if evidenceMetadata.CallerIdentifierMetadata != nil {
						startLine := int(evidenceMetadata.CallerIdentifierMetadata.StartLine)
						endLine := int(evidenceMetadata.CallerIdentifierMetadata.EndLine)

						snippet, err := extractFileSnippet(
							signatureMatchResult.FilePath,
							startLine,
							endLine,
							r.config.SnippetBeforeLines,
							r.config.SnippetAfterLines,
							r.config.SnippetMaxBytes,
							r.config.SnippetMaxLineChars,
						)

						if err != nil && strings.TrimSpace(evidenceMetadata.CallerIdentifierContent) != "" {
							// Fall back to CallerIdentifierContent
							lines := strings.Split(evidenceMetadata.CallerIdentifierContent, "\n")
							snippetLines := make([]snippetLineData, len(lines))
							for i, line := range lines {
								snippetLines[i] = snippetLineData{
									LineNum:     startLine + i + 1,
									Content:     line,
									IsMatch:     true,
									IsTruncated: false,
								}
							}
							match.Snippet = &snippetInfo{
								Lines:             snippetLines,
								RawContent:        evidenceMetadata.CallerIdentifierContent,
								WasTruncated:      false,
								SourceUnavailable: false,
							}
						} else if err == nil {
							// Successfully extracted from file
							match.Snippet = &snippetInfo{
								Lines:             snippet.Lines,
								RawContent:        snippet.RawContent,
								WasTruncated:      snippet.WasTruncated,
								SourceUnavailable: false,
							}
						} else {
							// File unavailable and no fallback
							match.Snippet = &snippetInfo{
								Lines:             []snippetLineData{},
								RawContent:        "",
								WasTruncated:      false,
								SourceUnavailable: true,
							}
						}
					}

					fileOcc.Matches = append(fileOcc.Matches, match)
					sigMap[sigID].TotalMatches++
				}
			}

			sigMap[sigID].FileOccurrences = append(sigMap[sigID].FileOccurrences, fileOcc)
		}
	}

	// Convert map to slice and sort by signature ID
	result := make([]signatureDetail, 0, len(sigMap))
	for _, sig := range sigMap {
		result = append(result, *sig)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

func (r *MarkdownReporter) getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"join": func(arr []string, sep string) string {
			return strings.Join(arr, sep)
		},
		"basename": func(path string) string {
			return filepath.Base(path)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"inc": func(i int) int {
			return i + 1
		},
		"formatLocation": func(line snippetLineData) string {
			marker := "  "
			if line.IsMatch {
				marker = "> "
			}
			return fmt.Sprintf("%s%4d | %s", marker, line.LineNum, line.Content)
		},
	}
}
