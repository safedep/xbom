package test

import (
	"path/filepath"
	"testing"

	callgraphv1 "buf.build/gen/go/safedep/api/protocolbuffers/go/safedep/messages/code/callgraph/v1"
	"github.com/safedep/xbom/pkg/codeanalysis"
	"github.com/safedep/xbom/pkg/common"
	"github.com/safedep/xbom/pkg/signatures"
	_ "github.com/safedep/xbom/signatures" // Initialize embedded signatures
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type expectedSignatureMatch struct {
	signatureID string
	filePath    string // relative to fixture dir
	condition   string // e.g., "os/WriteFile"
}

type codeAnalysisTestCase struct {
	name            string
	fixtureDir      string
	signatureFilter func() ([]*callgraphv1.Signature, error)
	expectedMatches []expectedSignatureMatch
	minMatchCount   int
}

func TestCodeAnalysisE2E(t *testing.T) {
	testCases := []codeAnalysisTestCase{
		{
			name:       "Go capabilities detection",
			fixtureDir: "fixtures/test_go_capabilities",
			signatureFilter: func() ([]*callgraphv1.Signature, error) {
				// Load all Go-related signatures from lang/golang/
				return signatures.LoadSignatures("lang/golang", "", "")
			},
			expectedMatches: []expectedSignatureMatch{
				// Filesystem operations
				{signatureID: "golang.filesystem.write", filePath: "main.go", condition: "os/WriteFile"},
				{signatureID: "golang.filesystem.read", filePath: "main.go", condition: "os/ReadFile"},
				{signatureID: "golang.filesystem.delete", filePath: "main.go", condition: "os/Remove"},
				{signatureID: "golang.filesystem.mkdir", filePath: "main.go", condition: "os/Mkdir"},
				// Network operations
				{signatureID: "golang.network.http.client", filePath: "main.go", condition: "net/http/Get"},
				{signatureID: "golang.network.http.server", filePath: "main.go", condition: "net/http/ListenAndServe"},
				// Process operations
				{signatureID: "golang.process.exec", filePath: "main.go", condition: "os/exec/Command"},
				{signatureID: "golang.process.info", filePath: "main.go", condition: "os/Getpid"},
				// Environment operations
				{signatureID: "golang.environment.read", filePath: "main.go", condition: "os/Getenv"},
				{signatureID: "golang.environment.write", filePath: "main.go", condition: "os/Setenv"},
				// Crypto operations
				{signatureID: "golang.crypto.hash", filePath: "main.go", condition: "crypto/sha256/Sum256"},
				{signatureID: "golang.crypto.aes", filePath: "main.go", condition: "crypto/aes/NewCipher"},
				// Database operations
				{signatureID: "golang.database.sql", filePath: "main.go", condition: "database/sql/Open"},
			},
			minMatchCount: 13,
		},
		{
			name:       "Python capabilities detection",
			fixtureDir: "fixtures/test_python_capabilities",
			signatureFilter: func() ([]*callgraphv1.Signature, error) {
				// Load all Python-related signatures from lang/python/
				return signatures.LoadSignatures("lang/python", "", "")
			},
			expectedMatches: []expectedSignatureMatch{
				// Filesystem operations
				{signatureID: "python.filesystem.delete", filePath: "main.py", condition: "os.remove"},
				{signatureID: "python.filesystem.mkdir", filePath: "main.py", condition: "os.mkdir"},
				// Network operations
				{signatureID: "python.network.http.client", filePath: "main.py", condition: "urllib.request.urlopen"},
				// Environment operations
				{signatureID: "python.environment.read", filePath: "main.py", condition: "os.getenv"},
				// Process operations
				{signatureID: "python.process.exec", filePath: "main.py", condition: "subprocess.run"},
				{signatureID: "python.process.info", filePath: "main.py", condition: "os.getpid"},
				// Database operations
				{signatureID: "python.database.sql", filePath: "main.py", condition: "sqlite3.connect"},
			},
			minMatchCount: 7,
		},
		{
			name:       "JavaScript capabilities detection",
			fixtureDir: "fixtures/test_javascript_capabilities",
			signatureFilter: func() ([]*callgraphv1.Signature, error) {
				// Load all JavaScript-related signatures from lang/javascript/
				return signatures.LoadSignatures("lang/javascript", "", "")
			},
			expectedMatches: []expectedSignatureMatch{
				// Filesystem operations
				{signatureID: "javascript.filesystem.write", filePath: "main.js", condition: "fs/writeFileSync"},
				{signatureID: "javascript.filesystem.read", filePath: "main.js", condition: "fs/readFileSync"},
				{signatureID: "javascript.filesystem.delete", filePath: "main.js", condition: "fs/unlinkSync"},
				{signatureID: "javascript.filesystem.mkdir", filePath: "main.js", condition: "fs/mkdirSync"},
				// Network operations
				{signatureID: "javascript.network.http.client", filePath: "main.js", condition: "http/get"},
				{signatureID: "javascript.network.http.server", filePath: "main.js", condition: "http/createServer"},
				// Process operations
				{signatureID: "javascript.process.exec", filePath: "main.js", condition: "child_process/exec"},
				// Crypto operations
				{signatureID: "javascript.crypto.hash", filePath: "main.js", condition: "crypto/createHash"},
			},
			minMatchCount: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load signatures using the filter function
			signaturesToMatch, err := tc.signatureFilter()
			require.NoError(t, err, "Failed to load signatures")
			require.NotEmpty(t, signaturesToMatch, "No signatures loaded")

			// Determine absolute path to fixture
			fixturePath, err := filepath.Abs(tc.fixtureDir)
			require.NoError(t, err, "Failed to get absolute path for fixture")

			// Create workflow with minimal config (no reporters, no callbacks)
			workflow := codeanalysis.NewCodeAnalysisWorkflow(
				codeanalysis.CodeAnalysisWorkflowConfig{
					Tool: common.ToolMetadata{
						Name:    "xbom-test",
						Version: "test",
					},
					SourcePath:        fixturePath,
					SignaturesToMatch: signaturesToMatch,
					Callbacks:         codeanalysis.CodeAnalysisCallbackRegistry{},
				},
				nil, // No reporters needed for testing
			)

			// Execute the analysis
			findings, err := workflow.Execute()
			require.NoError(t, err, "Code analysis workflow failed")
			require.NotNil(t, findings, "Findings should not be nil")

			// Validate minimum match count
			totalMatches := 0
			for _, matches := range findings.SignatureWiseMatchResults {
				totalMatches += len(matches)
			}
			assert.GreaterOrEqual(t, totalMatches, tc.minMatchCount,
				"Expected at least %d matches, got %d", tc.minMatchCount, totalMatches)

			// Validate each expected match
			for _, expected := range tc.expectedMatches {
				t.Run(expected.signatureID, func(t *testing.T) {
					matches, found := findings.SignatureWiseMatchResults[expected.signatureID]
					assert.True(t, found, "Signature %s not found in results", expected.signatureID)
					assert.NotEmpty(t, matches, "No matches for signature %s", expected.signatureID)

					// Check if at least one match has the expected file path and condition
					hasMatch := false
					for _, match := range matches {
						// Check file path (basename match is sufficient)
						if filepath.Base(match.FilePath) == expected.filePath {
							// Check if any matched condition has the expected value
							for _, cond := range match.MatchedConditions {
								if cond.Condition != nil && cond.Condition.Value == expected.condition {
									hasMatch = true
									break
								}
							}
						}
						if hasMatch {
							break
						}
					}

					assert.True(t, hasMatch,
						"Expected match for signature %s with file %s and condition %s not found",
						expected.signatureID, expected.filePath, expected.condition)
				})
			}
		})
	}
}
