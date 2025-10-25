package reporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	callgraphv1 "buf.build/gen/go/safedep/api/protocolbuffers/go/safedep/messages/code/callgraph/v1"
	"github.com/safedep/code/core"
	"github.com/safedep/code/plugin/callgraph"
	"github.com/safedep/xbom/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMarkdownReporter_ConfigDefaults(t *testing.T) {
	tests := []struct {
		name           string
		inputConfig    MarkdownReporterConfig
		expectedConfig MarkdownReporterConfig
	}{
		{
			name: "all defaults applied",
			inputConfig: MarkdownReporterConfig{
				OutputPath: "test.md",
			},
			expectedConfig: MarkdownReporterConfig{
				OutputPath:            "test.md",
				SnippetBeforeLines:    3,
				SnippetAfterLines:     3,
				SnippetMaxBytes:       5120,
				SnippetMaxLineChars:   500,
				ShowExecutiveSummary:  true,
				ShowStatistics:        true,
				ShowTopSignatures:     true,
				ShowLanguageBreakdown: true,
				ShowDetailedFindings:  true,
			},
		},
		{
			name: "custom snippet config",
			inputConfig: MarkdownReporterConfig{
				OutputPath:          "test.md",
				SnippetBeforeLines:  5,
				SnippetAfterLines:   5,
				SnippetMaxBytes:     10240,
				SnippetMaxLineChars: 1000,
			},
			expectedConfig: MarkdownReporterConfig{
				OutputPath:            "test.md",
				SnippetBeforeLines:    5,
				SnippetAfterLines:     5,
				SnippetMaxBytes:       10240,
				SnippetMaxLineChars:   1000,
				ShowExecutiveSummary:  true,
				ShowStatistics:        true,
				ShowTopSignatures:     true,
				ShowLanguageBreakdown: true,
				ShowDetailedFindings:  true,
			},
		},
		{
			name: "partial section visibility",
			inputConfig: MarkdownReporterConfig{
				OutputPath:           "test.md",
				ShowExecutiveSummary: true,
				ShowStatistics:       true,
			},
			expectedConfig: MarkdownReporterConfig{
				OutputPath:            "test.md",
				SnippetBeforeLines:    3,
				SnippetAfterLines:     3,
				SnippetMaxBytes:       5120,
				SnippetMaxLineChars:   500,
				ShowExecutiveSummary:  true,
				ShowStatistics:        true,
				ShowTopSignatures:     false,
				ShowLanguageBreakdown: false,
				ShowDetailedFindings:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter, err := NewMarkdownReporter(tt.inputConfig)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConfig, reporter.config)
		})
	}
}

func TestMarkdownReporter_Name(t *testing.T) {
	reporter, err := NewMarkdownReporter(MarkdownReporterConfig{OutputPath: "test.md"})
	require.NoError(t, err)
	assert.Equal(t, "markdown", reporter.Name())
}

func TestMarkdownReporter_RecordCodeAnalysisFindings(t *testing.T) {
	tests := []struct {
		name               string
		findings           *common.CodeAnalysisFindings
		expectedTotalCount int
		expectedFileCount  int
		expectedLangCount  int
		expectedSigCount   int
	}{
		{
			name: "single finding",
			findings: &common.CodeAnalysisFindings{
				SignatureWiseMatchResults: map[string][]common.EnrichedSignatureMatchResult{
					"sig1": {
						{
							SignatureMatchResult: callgraph.SignatureMatchResult{
								FilePath: "/test/file1.go",
								MatchedSignature: &callgraphv1.Signature{
									Id:          "sig1",
									Description: "Test signature 1",
									Tags:        []string{"test"},
								},
								MatchedLanguageCode: core.LanguageCodeGo,
								MatchedConditions: []callgraph.MatchedCondition{
									{
										Condition: &callgraphv1.Signature_LanguageMatcher_SignatureCondition{
											Type:  "pattern",
											Value: "test.*pattern",
										},
										Evidences: []callgraph.MatchedEvidence{
											{}, // Single evidence
										},
									},
								},
							},
						},
					},
				},
			},
			expectedTotalCount: 1,
			expectedFileCount:  1,
			expectedLangCount:  1,
			expectedSigCount:   1,
		},
		{
			name: "multiple findings across files and languages",
			findings: &common.CodeAnalysisFindings{
				SignatureWiseMatchResults: map[string][]common.EnrichedSignatureMatchResult{
					"sig1": {
						{
							SignatureMatchResult: callgraph.SignatureMatchResult{
								FilePath: "/test/file1.go",
								MatchedSignature: &callgraphv1.Signature{
									Id:          "sig1",
									Description: "Test signature 1",
								},
								MatchedLanguageCode: core.LanguageCodeGo,
								MatchedConditions: []callgraph.MatchedCondition{
									{
										Condition: &callgraphv1.Signature_LanguageMatcher_SignatureCondition{},
										Evidences: []callgraph.MatchedEvidence{{}, {}}, // 2 evidences
									},
								},
							},
						},
					},
					"sig2": {
						{
							SignatureMatchResult: callgraph.SignatureMatchResult{
								FilePath: "/test/file2.py",
								MatchedSignature: &callgraphv1.Signature{
									Id:          "sig2",
									Description: "Test signature 2",
								},
								MatchedLanguageCode: core.LanguageCodePython,
								MatchedConditions: []callgraph.MatchedCondition{
									{
										Condition: &callgraphv1.Signature_LanguageMatcher_SignatureCondition{},
										Evidences: []callgraph.MatchedEvidence{{}}, // 1 evidence
									},
								},
							},
						},
					},
				},
			},
			expectedTotalCount: 3, // 2 + 1
			expectedFileCount:  2,
			expectedLangCount:  2,
			expectedSigCount:   2,
		},
		{
			name:               "empty findings",
			findings:           &common.CodeAnalysisFindings{SignatureWiseMatchResults: map[string][]common.EnrichedSignatureMatchResult{}},
			expectedTotalCount: 0,
			expectedFileCount:  0,
			expectedLangCount:  0,
			expectedSigCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter, err := NewMarkdownReporter(MarkdownReporterConfig{OutputPath: "test.md"})
			require.NoError(t, err)

			err = reporter.RecordCodeAnalysisFindings(tt.findings)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedTotalCount, reporter.statistics.totalFindings)
			assert.Equal(t, tt.expectedFileCount, len(reporter.statistics.filesAffected))
			assert.Equal(t, tt.expectedLangCount, len(reporter.statistics.languageCounts))
			assert.Equal(t, tt.expectedSigCount, len(reporter.statistics.signatureCounts))
		})
	}
}

func TestMarkdownReporter_PrepareStatistics(t *testing.T) {
	reporter, _ := NewMarkdownReporter(MarkdownReporterConfig{OutputPath: "test.md"})
	reporter.statistics.totalFindings = 10
	reporter.statistics.filesAffected = map[string]bool{"file1": true, "file2": true}
	reporter.statistics.languageCounts = map[string]int{"go": 5, "python": 5}
	reporter.statistics.signatureCounts = map[string]int{"sig1": 7, "sig2": 3}

	stats := reporter.prepareStatistics()

	assert.Equal(t, 10, stats["TotalFindings"])
	assert.Equal(t, 2, stats["UniqueSignatures"])
	assert.Equal(t, 2, stats["FilesAffected"])
	assert.Equal(t, 2, stats["LanguagesDetected"])
}

func TestMarkdownReporter_PrepareTopSignatures(t *testing.T) {
	tests := []struct {
		name            string
		signatureCounts map[string]int
		expectedTop     []string // Expected signature IDs in order
	}{
		{
			name: "multiple signatures sorted by count",
			signatureCounts: map[string]int{
				"sig1": 5,
				"sig2": 10,
				"sig3": 3,
				"sig4": 8,
			},
			expectedTop: []string{"sig2", "sig4", "sig1", "sig3"},
		},
		{
			name: "more than 10 signatures - returns top 10",
			signatureCounts: map[string]int{
				"sig1":  1,
				"sig2":  2,
				"sig3":  3,
				"sig4":  4,
				"sig5":  5,
				"sig6":  6,
				"sig7":  7,
				"sig8":  8,
				"sig9":  9,
				"sig10": 10,
				"sig11": 11,
				"sig12": 12,
			},
			expectedTop: []string{"sig12", "sig11", "sig10", "sig9", "sig8", "sig7", "sig6", "sig5", "sig4", "sig3"},
		},
		{
			name:            "empty signatures",
			signatureCounts: map[string]int{},
			expectedTop:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter, _ := NewMarkdownReporter(MarkdownReporterConfig{OutputPath: "test.md"})
			reporter.statistics.signatureCounts = tt.signatureCounts

			topSigs := reporter.prepareTopSignatures()

			actualIDs := make([]string, len(topSigs))
			for i, sig := range topSigs {
				actualIDs[i] = sig["ID"].(string)
				assert.Equal(t, i+1, sig["Rank"])
			}

			assert.Equal(t, tt.expectedTop, actualIDs)
		})
	}
}

func TestMarkdownReporter_PrepareLanguageBreakdown(t *testing.T) {
	reporter, _ := NewMarkdownReporter(MarkdownReporterConfig{OutputPath: "test.md"})
	reporter.statistics.languageCounts = map[string]int{
		"python": 10,
		"go":     5,
		"java":   3,
	}

	breakdown := reporter.prepareLanguageBreakdown()

	// Should be sorted alphabetically
	assert.Len(t, breakdown, 3)
	assert.Equal(t, "go", breakdown[0]["Language"])
	assert.Equal(t, 5, breakdown[0]["Count"])
	assert.Equal(t, "java", breakdown[1]["Language"])
	assert.Equal(t, 3, breakdown[1]["Count"])
	assert.Equal(t, "python", breakdown[2]["Language"])
	assert.Equal(t, 10, breakdown[2]["Count"])
}

func TestMarkdownReporter_TemplateFunctions(t *testing.T) {
	reporter, _ := NewMarkdownReporter(MarkdownReporterConfig{OutputPath: "test.md"})
	funcs := reporter.getTemplateFuncs()

	t.Run("join function", func(t *testing.T) {
		joinFunc := funcs["join"].(func([]string, string) string)
		result := joinFunc([]string{"a", "b", "c"}, ", ")
		assert.Equal(t, "a, b, c", result)
	})

	t.Run("basename function", func(t *testing.T) {
		basenameFunc := funcs["basename"].(func(string) string)
		result := basenameFunc("/path/to/file.go")
		assert.Equal(t, "file.go", result)
	})

	t.Run("lower function", func(t *testing.T) {
		lowerFunc := funcs["lower"].(func(string) string)
		result := lowerFunc("HELLO")
		assert.Equal(t, "hello", result)
	})

	t.Run("inc function", func(t *testing.T) {
		incFunc := funcs["inc"].(func(int) int)
		result := incFunc(5)
		assert.Equal(t, 6, result)
	})

	t.Run("formatLocation function", func(t *testing.T) {
		formatFunc := funcs["formatLocation"].(func(snippetLineData) string)

		// Test match line
		matchLine := snippetLineData{LineNum: 10, Content: "test content", IsMatch: true}
		result := formatFunc(matchLine)
		assert.Contains(t, result, ">")
		assert.Contains(t, result, "10")
		assert.Contains(t, result, "test content")

		// Test non-match line
		normalLine := snippetLineData{LineNum: 11, Content: "normal line", IsMatch: false}
		result = formatFunc(normalLine)
		assert.Contains(t, result, "  ")
		assert.Contains(t, result, "11")
		assert.Contains(t, result, "normal line")
	})
}

func TestMarkdownReporter_GenerateReport_WithFindings(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "report.md")

	reporter, err := NewMarkdownReporter(MarkdownReporterConfig{
		OutputPath: outputPath,
	})
	require.NoError(t, err)

	// Create test findings
	findings := &common.CodeAnalysisFindings{
		SignatureWiseMatchResults: map[string][]common.EnrichedSignatureMatchResult{
			"test-sig-1": {
				{
					SignatureMatchResult: callgraph.SignatureMatchResult{
						FilePath: "/test/sample.go",
						MatchedSignature: &callgraphv1.Signature{
							Id:          "test-sig-1",
							Description: "Test signature for demonstration",
							Tags:        []string{"test", "demo"},
						},
						MatchedLanguageCode: core.LanguageCodeGo,
						MatchedConditions: []callgraph.MatchedCondition{
							{
								Condition: &callgraphv1.Signature_LanguageMatcher_SignatureCondition{
									Type:  "pattern",
									Value: "fmt.Println",
								},
								Evidences: []callgraph.MatchedEvidence{{}},
							},
						},
					},
				},
			},
		},
	}

	err = reporter.RecordCodeAnalysisFindings(findings)
	require.NoError(t, err)

	err = reporter.Finish()
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, outputPath)

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Verify markdown structure exists
	assert.True(t, strings.HasPrefix(contentStr, "# "), "Should start with H1 header")
	assert.Contains(t, contentStr, "---", "Should contain horizontal rules")
	assert.Contains(t, contentStr, "|", "Should contain tables")

	// Verify actual data is present (not template text)
	assert.Contains(t, contentStr, "test-sig-1", "Should contain signature ID")
	assert.Contains(t, contentStr, "Test signature for demonstration", "Should contain signature description")
	assert.Contains(t, contentStr, "test, demo", "Should contain tags")
	assert.Contains(t, contentStr, "sample.go", "Should contain filename")
}

func TestMarkdownReporter_GenerateReport_NoFindings(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "report.md")

	reporter, err := NewMarkdownReporter(MarkdownReporterConfig{
		OutputPath: outputPath,
	})
	require.NoError(t, err)

	// Record empty findings
	findings := &common.CodeAnalysisFindings{
		SignatureWiseMatchResults: map[string][]common.EnrichedSignatureMatchResult{},
	}

	err = reporter.RecordCodeAnalysisFindings(findings)
	require.NoError(t, err)

	err = reporter.Finish()
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, outputPath)

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Verify report was generated
	assert.True(t, strings.HasPrefix(contentStr, "# "), "Should start with H1 header")
	assert.Greater(t, len(contentStr), 100, "Report should have content")

	// Should NOT contain statistics sections when there are no findings (conditional rendering)
	assert.NotContains(t, contentStr, "| Total Findings", "Should not show statistics table")
	assert.NotContains(t, contentStr, "| Rank | Signature ID", "Should not show top signatures table")

	// Verify the report indicates no findings were found (don't check exact template text)
	containsNoFindingsIndicator := strings.Contains(contentStr, "No Findings") ||
		strings.Contains(contentStr, "no signature matches") ||
		strings.Contains(contentStr, "0") // At least should show 0 findings somewhere
	assert.True(t, containsNoFindingsIndicator, "Report should indicate no findings were found")
}

func TestMarkdownReporter_GenerateReport_SectionVisibility(t *testing.T) {
	tests := []struct {
		name                      string
		config                    MarkdownReporterConfig
		expectedContentIndicators []string // Check for data/content presence, not exact headers
		notExpectedIndicators     []string // Check for data absence
	}{
		{
			name: "only statistics enabled",
			config: MarkdownReporterConfig{
				ShowExecutiveSummary:  false,
				ShowStatistics:        true,
				ShowTopSignatures:     false,
				ShowLanguageBreakdown: false,
				ShowDetailedFindings:  false,
			},
			// Check for statistics table content, not header text
			expectedContentIndicators: []string{"| Total Findings", "| Unique Signatures Matched"},
			// Check that other section content is absent
			notExpectedIndicators: []string{
				"| Rank | Signature ID",         // Top signatures table
				"### 1. Signature:",              // Detailed findings
				"This report provides a comprehensive", // Executive summary text
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, "report.md")

			tt.config.OutputPath = outputPath
			reporter, err := NewMarkdownReporter(tt.config)
			require.NoError(t, err)

			// Add findings to test section visibility
			findings := &common.CodeAnalysisFindings{
				SignatureWiseMatchResults: map[string][]common.EnrichedSignatureMatchResult{
					"sig1": {
						{
							SignatureMatchResult: callgraph.SignatureMatchResult{
								FilePath: "/test/file.go",
								MatchedSignature: &callgraphv1.Signature{
									Id:          "sig1",
									Description: "Test",
								},
								MatchedLanguageCode: core.LanguageCodeGo,
								MatchedConditions: []callgraph.MatchedCondition{
									{
										Condition: &callgraphv1.Signature_LanguageMatcher_SignatureCondition{},
										Evidences: []callgraph.MatchedEvidence{{}},
									},
								},
							},
						},
					},
				},
			}

			err = reporter.RecordCodeAnalysisFindings(findings)
			require.NoError(t, err)

			err = reporter.Finish()
			require.NoError(t, err)

			content, err := os.ReadFile(outputPath)
			require.NoError(t, err)
			contentStr := string(content)

			// Check expected content is present (focus on data, not template headers)
			for _, indicator := range tt.expectedContentIndicators {
				assert.Contains(t, contentStr, indicator, "Expected content indicator %q not found", indicator)
			}

			// Check unexpected content is not present
			for _, indicator := range tt.notExpectedIndicators {
				assert.NotContains(t, contentStr, indicator, "Unexpected content indicator %q found", indicator)
			}
		})
	}
}

func TestMarkdownTemplate_ValidMarkdownStructure(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "report.md")

	reporter, err := NewMarkdownReporter(MarkdownReporterConfig{
		OutputPath: outputPath,
	})
	require.NoError(t, err)

	findings := &common.CodeAnalysisFindings{
		SignatureWiseMatchResults: map[string][]common.EnrichedSignatureMatchResult{
			"sig1": {
				{
					SignatureMatchResult: callgraph.SignatureMatchResult{
						FilePath: "/test/file.py",
						MatchedSignature: &callgraphv1.Signature{
							Id:          "sig1",
							Description: "Test signature",
							Tags:        []string{"python", "test"},
						},
						MatchedLanguageCode: core.LanguageCodePython,
						MatchedConditions: []callgraph.MatchedCondition{
							{
								Condition: &callgraphv1.Signature_LanguageMatcher_SignatureCondition{
									Type:  "pattern",
									Value: "import os",
								},
								Evidences: []callgraph.MatchedEvidence{{}},
							},
						},
					},
				},
			},
		},
	}

	err = reporter.RecordCodeAnalysisFindings(findings)
	require.NoError(t, err)

	err = reporter.Finish()
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Verify basic markdown structure
	assert.True(t, strings.HasPrefix(contentStr, "# "), "Should start with H1 header")
	assert.Contains(t, contentStr, "---", "Should contain horizontal rules for sections")
	assert.Contains(t, contentStr, "|", "Should contain tables")

	// Verify actual data is present (not template text)
	assert.Contains(t, contentStr, "sig1", "Should contain signature ID")
	assert.Contains(t, contentStr, "Test signature", "Should contain signature description")
	assert.Contains(t, contentStr, "python", "Should contain language")
	assert.Contains(t, contentStr, "/test/file.py", "Should contain file path")
}
