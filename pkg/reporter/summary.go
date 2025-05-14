package reporter

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/safedep/xbom/pkg/codeanalysis"
)

func SummariseCodeAnalysisFindings(codeAnalysisFindings *codeanalysis.CodeAnalysisFindings) error {
	sigTable := table.NewWriter()
	sigTable.SetOutputMirror(os.Stdout)
	sigTable.SetStyle(table.StyleRounded)

	sigTable.AppendHeader(table.Row{"Signature", "Language", "Condition", "Evidence file", "Location"})
	sigTable.SetTitle("Matched Signatures")

	sigTable.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:    0,
			AutoMerge: true,
			WidthMax:  20,
		},
		{
			Number:   1,
			WidthMax: 10,
		},
		{
			Number:   3,
			WidthMax: 25,
		},
		{
			Number:   4,
			WidthMax: 25,
		},
	})

	for _, signatureResults := range codeAnalysisFindings.SignatureWiseMatchResults {
		for _, match := range signatureResults {
			for _, condition := range match.MatchedConditions {
				for _, evidence := range condition.Evidences {
					evidenceDetailString := "Unknown"
					evidenceContent, exists := evidence.Metadata()
					if exists {
						evidenceDetailString = fmt.Sprintf("L%d:%d to\nL%d:%d",
							evidenceContent.StartLine, evidenceContent.StartColumn,
							evidenceContent.EndLine, evidenceContent.EndColumn)
					}

					conditionLocationString := fmt.Sprintf("%s: \n%s", condition.Condition.Type, condition.Condition.Value)
					sigTable.AppendRow(table.Row{
						match.MatchedSignature.Id,
						match.MatchedLanguageCode,
						conditionLocationString,
						match.FilePath,
						evidenceDetailString,
					})
					sigTable.AppendSeparator()
				}
			}
		}
	}

	sigTable.Render()

	return nil
}
