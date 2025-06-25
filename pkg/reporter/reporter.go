package reporter

import (
	"github.com/safedep/xbom/pkg/common"
)

type Reporter interface {
	Name() string

	// Feed collected data to reporting module
	RecordCodeAnalysisFindings(findings *common.CodeAnalysisFindings) error

	// Inform reporting module to finalise (e.g. write report to file)
	Finish() error
}
