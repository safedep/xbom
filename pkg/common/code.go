package common

import "github.com/safedep/code/plugin/callgraph"

type EnrichedSignatureMatchResult struct {
	callgraph.SignatureMatchResult
	TreeData *[]byte
}

type CodeAnalysisFindings struct {
	SignatureWiseMatchResults map[string][]EnrichedSignatureMatchResult
}
