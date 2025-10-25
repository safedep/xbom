package reporter

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/safedep/dry/log"
	"github.com/safedep/xbom/pkg/common"
)

type HTMLReporterConfig struct {
	HtmlReportPath      string // Path to save the HTML report
	SnippetBeforeLines  int    // Number of context lines to show before match (default: 3)
	SnippetAfterLines   int    // Number of context lines to show after match (default: 3)
	SnippetMaxBytes     int    // Max total bytes for snippet (default: 5120 = 5KB)
	SnippetMaxLineChars int    // Max characters per line (default: 500)
}

type HTMLReporter struct {
	config     HTMLReporterConfig
	visualiser *HTMLVisualiser
}

var _ Reporter = (*HTMLReporter)(nil)

func NewHTMLReporter(config HTMLReporterConfig) (*HTMLReporter, error) {
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

	return &HTMLReporter{
		config:     config,
		visualiser: NewHTMLVisualiser([]string{"Signature ID", "Description", "Tags"}),
	}, nil
}

func (r *HTMLReporter) Name() string {
	return "html"
}

func (r *HTMLReporter) RecordCodeAnalysisFindings(codeAnalysisFindings *common.CodeAnalysisFindings) error {
	sigRows := map[string]map[string]interface{}{}

	for _, signatureResults := range codeAnalysisFindings.SignatureWiseMatchResults {
		for _, signatureMatchResult := range signatureResults {
			sig := signatureMatchResult.MatchedSignature
			sigId := sig.Id
			desc := sig.Description
			tags := strings.Join(sig.Tags, ", ")

			fileMap := make(map[string]map[string]interface{})

			for _, condition := range signatureMatchResult.MatchedConditions {
				for _, evidence := range condition.Evidences {
					evidenceMetadata := evidence.Metadata(signatureMatchResult.TreeData)

					key := signatureMatchResult.FilePath + "|" + string(signatureMatchResult.MatchedLanguageCode)
					if _, ok := fileMap[key]; !ok {
						fileMap[key] = map[string]interface{}{
							"File":     signatureMatchResult.FilePath,
							"Language": string(signatureMatchResult.MatchedLanguageCode),
							"Matches":  []map[string]interface{}{},
						}
					}

					conditionValueString := fmt.Sprintf("%s - %s", condition.Condition.Type, strings.ReplaceAll(condition.Condition.Value, "\n", " "))

					// Create a match object with occurrence and snippet
					match := map[string]interface{}{
						"Occurrence": conditionValueString,
						"Snippet":    nil,
					}

					// Add snippet if available
					if evidenceMetadata.CallerIdentifierMetadata != nil {
						startLine := int(evidenceMetadata.CallerIdentifierMetadata.StartLine)
						endLine := int(evidenceMetadata.CallerIdentifierMetadata.EndLine)

						// Try to extract snippet from file with context
						snippet, err := extractFileSnippet(
							signatureMatchResult.FilePath,
							startLine,
							endLine,
							r.config.SnippetBeforeLines,
							r.config.SnippetAfterLines,
							r.config.SnippetMaxBytes,
							r.config.SnippetMaxLineChars,
						)

						// If file extraction fails, fall back to CallerIdentifierContent
						if err != nil && strings.TrimSpace(evidenceMetadata.CallerIdentifierContent) != "" {
							lines := strings.Split(evidenceMetadata.CallerIdentifierContent, "\n")
							snippetLines := make([]map[string]interface{}, len(lines))

							for i, line := range lines {
								snippetLines[i] = map[string]interface{}{
									"LineNum":     startLine + i + 1,
									"Content":     line,
									"IsMatch":     true,
									"IsTruncated": false,
								}
							}

							match["Snippet"] = map[string]interface{}{
								"Lines":             snippetLines,
								"RawContent":        evidenceMetadata.CallerIdentifierContent,
								"WasTruncated":      false,
								"SourceUnavailable": false,
							}
						} else if err == nil {
							// Successfully extracted snippet from file
							snippetLines := make([]map[string]interface{}, len(snippet.Lines))
							for i, line := range snippet.Lines {
								snippetLines[i] = map[string]interface{}{
									"LineNum":     line.LineNum,
									"Content":     line.Content,
									"IsMatch":     line.IsMatch,
									"IsTruncated": line.IsTruncated,
								}
							}

							match["Snippet"] = map[string]interface{}{
								"Lines":             snippetLines,
								"RawContent":        snippet.RawContent,
								"WasTruncated":      snippet.WasTruncated,
								"SourceUnavailable": false,
							}
						} else {
							// File unavailable and no fallback content
							match["Snippet"] = map[string]interface{}{
								"Lines":             []map[string]interface{}{},
								"RawContent":        "",
								"WasTruncated":      false,
								"SourceUnavailable": true,
							}
						}
					}

					// Add the match to the file's matches array
					fileMap[key]["Matches"] = append(
						fileMap[key]["Matches"].([]map[string]interface{}),
						match,
					)
				}
			}

			if _, ok := sigRows[sigId]; !ok {
				sigRows[sigId] = map[string]interface{}{
					"Signature ID":    sigId,
					"Description":     desc,
					"Tags":            tags,
					"FileOccurrences": []map[string]interface{}{},
				}
			}
			existing := sigRows[sigId]["FileOccurrences"].([]map[string]interface{})
			for _, v := range fileMap {
				existing = append(existing, v)
			}
			sigRows[sigId]["FileOccurrences"] = existing
		}
	}

	for _, row := range sigRows {
		r.visualiser.AddRow(row)
	}

	return nil
}

func (r *HTMLReporter) Finish() error {
	if r.visualiser == nil {
		return fmt.Errorf("visualiser is not initialized correctly")
	}

	if err := r.visualiser.GenerateHtmlFile(r.config.HtmlReportPath); err != nil {
		return fmt.Errorf("failed to finish HTML report: %w", err)
	}

	fmt.Println("🔗 You can view the HTML report at:", r.config.HtmlReportPath)

	return nil
}

//go:embed templates/report.html
var templateFS embed.FS

// getHTMLTemplate returns the HTML template content from the embedded file
func getHTMLTemplate() (string, error) {
	data, err := templateFS.ReadFile("templates/report.html")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// HTMLVisualiser builds and writes an interactive HTML report
type HTMLVisualiser struct {
	headers []string
	rows    []map[string]interface{}
}

func NewHTMLVisualiser(headers []string) *HTMLVisualiser {
	return &HTMLVisualiser{
		headers: headers,
		rows:    []map[string]interface{}{},
	}
}

func (hv *HTMLVisualiser) AddRow(row map[string]interface{}) {
	hv.rows = append(hv.rows, row)
}

func (hv *HTMLVisualiser) GenerateHtmlFile(htmlPath string) error {
	// Map of popular languages to their CDN icon links
	languageIconMap := map[string]string{
		"python":     "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/python/python-original.svg",
		"javascript": "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/javascript/javascript-original.svg",
		"go":         "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original.svg",
		"java":       "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/java/java-original.svg",
		"c":          "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/c/c-original.svg",
		"cpp":        "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/cplusplus/cplusplus-original.svg",
		"ruby":       "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/ruby/ruby-original.svg",
	}

	htmlTemplate, err := getHTMLTemplate()
	if err != nil {
		return fmt.Errorf("failed to load HTML template: %v", err)
	}

	t := template.Must(template.New("report").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).Parse(htmlTemplate))

	f, err := os.Create(htmlPath)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("failed to close HTML report file: %v", err)
		}
	}()

	headers := []string{"Signature_ID", "Description", "Tags"}
	var rows []map[string]interface{}
	tagSet := make(map[string]struct{})
	for _, row := range hv.rows {
		rows = append(rows, map[string]interface{}{
			"Signature_ID":    row["Signature ID"],
			"Description":     row["Description"],
			"Tags":            row["Tags"],
			"FileOccurrences": row["FileOccurrences"],
		})

		tags := strings.Split(row["Tags"].(string), ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			tagSet[tag] = struct{}{}
		}
	}

	var uniqueTags []string
	for tag := range tagSet {
		if tag != "" {
			uniqueTags = append(uniqueTags, tag)
		}
	}

	return t.Execute(f, map[string]interface{}{
		"Headers":         headers,
		"Rows":            rows,
		"UniqueTags":      uniqueTags,
		"LanguageIconMap": languageIconMap,
	})
}
