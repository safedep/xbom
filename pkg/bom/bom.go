package bom

import "github.com/safedep/xbom/pkg/codeanalysis"

type BomGenerator interface {
	RecordCodeAnalysisFindings(findings *codeanalysis.CodeAnalysisFindings) error
	Finish() error
}
