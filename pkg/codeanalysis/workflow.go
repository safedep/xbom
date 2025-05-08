package codeanalysis

import (
	callgraphv1 "buf.build/gen/go/safedep/api/protocolbuffers/go/safedep/messages/code/callgraph/v1"
	"github.com/safedep/code/plugin/callgraph"
	"github.com/safedep/xbom/pkg/common"
)

type CodeAnalysisWorkflowConfig struct {
	Tool              common.ToolMetadata
	SourcePath        string
	SignaturesToMatch []*callgraphv1.Signature
	Callbacks         *CodeAnalysisCallbackRegistry
}

type CodeAnalysisFindings struct {
	SignatureMatchResults []callgraph.SignatureMatchResult
}
