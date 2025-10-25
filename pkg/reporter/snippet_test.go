package reporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFileSnippet(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	testContent := `package main

import "fmt"

func main() {
	fmt.Println("line 6")
	fmt.Println("line 7")
	fmt.Println("line 8")
	fmt.Println("line 9")
	fmt.Println("line 10")
}
`
	err := os.WriteFile(testFile, []byte(testContent), 0o644)
	require.NoError(t, err, "failed to create test file")

	// Create file with a very long line for truncation testing
	longLineFile := filepath.Join(tmpDir, "long.go")
	longLine := "package main\n\n// " + strings.Repeat("x", 600) + "\n"
	err = os.WriteFile(longLineFile, []byte(longLine), 0o644)
	require.NoError(t, err, "failed to create long line test file")

	tests := []struct {
		name         string
		filePath     string
		startLine    int
		endLine      int
		beforeLines  int
		afterLines   int
		maxBytes     int
		maxLineChars int
		wantErr      bool
		validate     func(t *testing.T, snippet snippetData, err error)
	}{
		{
			name:         "extract with context",
			filePath:     testFile,
			startLine:    6,
			endLine:      6,
			beforeLines:  2,
			afterLines:   2,
			maxBytes:     10000,
			maxLineChars: 500,
			wantErr:      false,
			validate: func(t *testing.T, snippet snippetData, err error) {
				assert.NoError(t, err)
				assert.False(t, snippet.WasTruncated)

				// Should have 5 lines total: 2 before + 1 match + 2 after
				assert.Len(t, snippet.Lines, 5)

				// Check line numbers
				assert.Equal(t, 5, snippet.Lines[0].LineNum) // line 5
				assert.Equal(t, 6, snippet.Lines[1].LineNum) // line 6
				assert.Equal(t, 7, snippet.Lines[2].LineNum) // line 7 (match)
				assert.Equal(t, 8, snippet.Lines[3].LineNum) // line 8
				assert.Equal(t, 9, snippet.Lines[4].LineNum) // line 9

				// Check IsMatch flag
				assert.False(t, snippet.Lines[0].IsMatch)
				assert.False(t, snippet.Lines[1].IsMatch)
				assert.True(t, snippet.Lines[2].IsMatch) // The matched line
				assert.False(t, snippet.Lines[3].IsMatch)
				assert.False(t, snippet.Lines[4].IsMatch)
			},
		},
		{
			name:         "truncate long lines",
			filePath:     longLineFile,
			startLine:    2,
			endLine:      2,
			beforeLines:  0,
			afterLines:   0,
			maxBytes:     10000,
			maxLineChars: 100,
			wantErr:      false,
			validate: func(t *testing.T, snippet snippetData, err error) {
				assert.NoError(t, err)
				assert.Len(t, snippet.Lines, 1)
				assert.True(t, snippet.Lines[0].IsTruncated)
				assert.Contains(t, snippet.Lines[0].Content, "... (truncated)")
			},
		},
		{
			name:         "truncate on max bytes",
			filePath:     testFile,
			startLine:    5,
			endLine:      5,
			beforeLines:  2,
			afterLines:   10,
			maxBytes:     50,
			maxLineChars: 500,
			wantErr:      false,
			validate: func(t *testing.T, snippet snippetData, err error) {
				assert.NoError(t, err)
				assert.True(t, snippet.WasTruncated)

				// Should have stopped before reaching all requested lines
				// Would be 13 without truncation (2 before + 1 + 10 after)
				assert.Less(t, len(snippet.Lines), 13)
			},
		},
		{
			name:         "handle file boundaries",
			filePath:     testFile,
			startLine:    0,
			endLine:      0,
			beforeLines:  5,
			afterLines:   2,
			maxBytes:     10000,
			maxLineChars: 500,
			wantErr:      false,
			validate: func(t *testing.T, snippet snippetData, err error) {
				assert.NoError(t, err)
				// Should start at line 1 even though we requested 5 lines before line 0
				assert.Equal(t, 1, snippet.Lines[0].LineNum)
			},
		},
		{
			name:         "handle missing file",
			filePath:     "/nonexistent/file.go",
			startLine:    0,
			endLine:      0,
			beforeLines:  1,
			afterLines:   1,
			maxBytes:     1000,
			maxLineChars: 500,
			wantErr:      true,
			validate: func(t *testing.T, snippet snippetData, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snippet, err := extractFileSnippet(
				tt.filePath,
				tt.startLine,
				tt.endLine,
				tt.beforeLines,
				tt.afterLines,
				tt.maxBytes,
				tt.maxLineChars,
			)

			if tt.validate != nil {
				tt.validate(t, snippet, err)
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
