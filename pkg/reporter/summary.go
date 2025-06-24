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
			WidthMax:  40,
		},
		{
			Number:   1,
			WidthMax: 10,
		},
		{
			Number:   3,
			WidthMax: 30,
		},
		{
			Number:   4,
			WidthMax: 30,
		},
	})

	for _, signatureResults := range codeAnalysisFindings.SignatureWiseMatchResults {
		for _, signatureMatchResult := range signatureResults {
			for _, condition := range signatureMatchResult.MatchedConditions {
				for _, evidence := range condition.Evidences {
					evidenceDetailString := "Unknown"
					evidenceMetadata := evidence.Metadata(signatureMatchResult.TreeData)
					if evidenceMetadata.CallerIdentifierMetadata != nil {
						evidenceDetailString = fmt.Sprintf(
							"L%d:%d to\nL%d:%d",
							evidenceMetadata.CallerIdentifierMetadata.StartLine+1,
							evidenceMetadata.CallerIdentifierMetadata.StartColumn+1,
							evidenceMetadata.CallerIdentifierMetadata.EndLine+1,
							evidenceMetadata.CallerIdentifierMetadata.EndColumn+1,
						)
					}

					conditionLocationString := fmt.Sprintf("%s: \n%s", condition.Condition.Type, condition.Condition.Value)
					sigTable.AppendRow(table.Row{
						signatureMatchResult.MatchedSignature.Id,
						signatureMatchResult.MatchedLanguageCode,
						conditionLocationString,
						signatureMatchResult.FilePath,
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
