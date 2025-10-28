package reporter

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func TestTemplateGeneratesValidHTML(t *testing.T) {
	htmlTemplate, err := getHTMLTemplate()
	assert.NoError(t, err, "failed to get HTML template")

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).Parse(htmlTemplate)
	assert.NoError(t, err, "failed to parse HTML template")

	// Minimal valid dummy data to test rendering
	data := map[string]interface{}{
		"Headers": []string{"Signature_ID", "Description", "Tags"},
		"Rows": []map[string]interface{}{
			{
				"Signature_ID": "sample-signature-1",
				"Description":  "Sample signature",
				"Tags":         "go,security",
				"FileOccurrences": []map[string]interface{}{
					{
						"File":     "main.go",
						"Language": "go",
						"Matches": []map[string]interface{}{
							{
								"Occurrence": "Some match",
								"Snippet": map[string]interface{}{
									"RawContent": "fmt.Println(\"Hello\")",
									"Lines": []map[string]interface{}{
										{"LineNum": 10, "Content": "fmt.Println(\"Hello\")"},
									},
								},
							},
						},
					},
				},
			},
		},
		"UniqueTags": []string{"go", "security"},
		"LanguageIconMap": map[string]string{
			"go": "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original.svg",
		},
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	assert.NoError(t, err, "failed to render HTML template")

	// Validate HTML syntax
	htmlNode, err := html.Parse(&buf)
	assert.NoError(t, err, "generated HTML is not valid")
	assert.NotNil(t, htmlNode, "HTML node should not be nil")

	doc, err := goquery.NewDocumentFromReader(&buf)
	assert.NoError(t, err, "failed to parse HTML with goquery")
	assert.NotNil(t, doc, "goquery document should not be nil")

	// Check if basic structure is present such as tags: <html>, <head>, and <body> are present
	requiredTags := []string{"html", "head", "body"}
	for _, tag := range requiredTags {
		assert.NotZero(t, doc.Find(tag).Length(), "Missing <%s> tag in HTML", tag)
	}
}

func TestTemplateWithEnhancedSnippetFields(t *testing.T) {
	htmlTemplate, err := getHTMLTemplate()
	assert.NoError(t, err, "failed to get HTML template")

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).Parse(htmlTemplate)
	assert.NoError(t, err, "failed to parse HTML template")

	// Test with new snippet fields
	data := map[string]interface{}{
		"Headers": []string{"Signature_ID", "Description", "Tags"},
		"Rows": []map[string]interface{}{
			{
				"Signature_ID": "test-sig-1",
				"Description":  "Test signature with enhanced snippet",
				"Tags":         "test",
				"FileOccurrences": []map[string]interface{}{
					{
						"File":     "test.go",
						"Language": "go",
						"Matches": []map[string]interface{}{
							{
								"Occurrence": "Test match",
								"Snippet": map[string]interface{}{
									"RawContent":    "line1\nline2\nline3",
									"WasTruncated":  true,
									"SourceUnavailable": false,
									"Lines": []map[string]interface{}{
										{"LineNum": 1, "Content": "line1", "IsMatch": false, "IsTruncated": false},
										{"LineNum": 2, "Content": "line2 very long...", "IsMatch": true, "IsTruncated": true},
										{"LineNum": 3, "Content": "line3", "IsMatch": false, "IsTruncated": false},
									},
								},
							},
							{
								"Occurrence": "Unavailable source",
								"Snippet": map[string]interface{}{
									"RawContent":          "",
									"WasTruncated":        false,
									"SourceUnavailable":   true,
									"Lines":               []map[string]interface{}{},
								},
							},
						},
					},
				},
			},
		},
		"UniqueTags": []string{"test"},
		"LanguageIconMap": map[string]string{
			"go": "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original.svg",
		},
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	assert.NoError(t, err, "failed to render HTML template with enhanced fields")

	htmlStr := buf.String()

	// Verify new fields are rendered
	assert.Contains(t, htmlStr, "⚠️ Snippet truncated due to size limit")
	assert.Contains(t, htmlStr, "⚠️ Source file unavailable")
	assert.Contains(t, htmlStr, "bg-yellow-900 bg-opacity-40") // IsMatch highlighting

	// Validate HTML syntax
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	assert.NoError(t, err, "failed to parse HTML with enhanced snippet fields")
	assert.NotNil(t, doc, "goquery document should not be nil")
}
