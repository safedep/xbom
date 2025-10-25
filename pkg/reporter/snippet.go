package reporter

import (
	"fmt"
	"os"
	"strings"
)

// snippetLineData represents a single line in a code snippet
type snippetLineData struct {
	LineNum     int
	Content     string
	IsMatch     bool
	IsTruncated bool
}

// snippetData represents the complete snippet information
type snippetData struct {
	Lines        []snippetLineData
	RawContent   string
	WasTruncated bool
}

// extractFileSnippet reads a file and extracts a code snippet with context around the match.
// It handles size limits to prevent excessive memory usage with minified or obfuscated code.
func extractFileSnippet(filePath string,
	startLine, endLine, beforeLines, afterLines, maxBytes, maxLineChars int,
) (snippetData, error) {
	result := snippetData{
		Lines:        []snippetLineData{},
		RawContent:   "",
		WasTruncated: false,
	}

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return result, fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Calculate the range of lines to extract
	// Note: startLine and endLine are 0-indexed from the metadata
	firstLine := max(startLine-beforeLines, 0)
	lastLine := min(endLine+afterLines, len(lines)-1)

	// Extract lines with size limits
	totalBytes := 0
	var rawContentBuilder strings.Builder

	for i := firstLine; i <= lastLine; i++ {
		line := lines[i]
		lineNum := i + 1 // Display line numbers as 1-indexed

		// Check if this line is part of the match
		isMatch := i >= startLine && i <= endLine

		// Truncate line if too long
		isTruncated := false
		if len(line) > maxLineChars {
			line = line[:maxLineChars] + "... (truncated)"
			isTruncated = true
		}

		// Check if adding this line would exceed max bytes
		lineBytes := len(line) + 1 // +1 for newline
		if totalBytes+lineBytes > maxBytes {
			result.WasTruncated = true
			break
		}

		result.Lines = append(result.Lines, snippetLineData{
			LineNum:     lineNum,
			Content:     line,
			IsMatch:     isMatch,
			IsTruncated: isTruncated,
		})

		rawContentBuilder.WriteString(line)
		rawContentBuilder.WriteString("\n")
		totalBytes += lineBytes
	}

	result.RawContent = rawContentBuilder.String()
	return result, nil
}
