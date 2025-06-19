package reporter

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/safedep/xbom/pkg/codeanalysis"
)

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

func (hv *HTMLVisualiser) Finish(htmlPath string) error {
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
	defer f.Close()

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
		uniqueTags = append(uniqueTags, tag)
	}

	return t.Execute(f, map[string]interface{}{
		"Headers":         headers,
		"Rows":            rows,
		"LanguageIconMap": languageIconMap,
	})
}

func VisualiseCodeAnalysisFindings(codeAnalysisFindings *codeanalysis.CodeAnalysisFindings, htmlPath string) error {
	hv := NewHTMLVisualiser([]string{"Signature ID", "Description", "Tags"})

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
						"Occurrence": fmt.Sprintf("%s", conditionValueString),
						"Snippet":    nil,
					}

					// Add snippet if available
					if evidenceMetadata.CallerIdentifierMetadata != nil && strings.TrimSpace(evidenceMetadata.CallerIdentifierContent) != "" {
						lines := strings.Split(evidenceMetadata.CallerIdentifierContent, "\n")
						snippetLines := make([]map[string]interface{}, len(lines))
						startLine := int(evidenceMetadata.CallerIdentifierMetadata.StartLine)

						for i, line := range lines {
							snippetLines[i] = map[string]interface{}{
								"LineNum": startLine + i + 1,
								"Content": line,
							}
						}

						match["Snippet"] = map[string]interface{}{
							"Lines": snippetLines,
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
		hv.AddRow(row)
	}

	return hv.Finish(htmlPath)
}
