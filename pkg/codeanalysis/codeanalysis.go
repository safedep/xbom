package codeanalysis

import (
	"context"
	"fmt"

	"github.com/safedep/code/core"
	"github.com/safedep/code/fs"
	"github.com/safedep/code/lang"
	"github.com/safedep/code/parser"
	"github.com/safedep/code/plugin"
	"github.com/safedep/code/plugin/callgraph"
	"github.com/safedep/xbom/pkg/common"
	"github.com/safedep/xbom/pkg/reporter"
)

type CodeAnalysisWorkflow struct {
	config    CodeAnalysisWorkflowConfig
	findings  common.CodeAnalysisFindings
	reporters []reporter.Reporter
}

func NewCodeAnalysisWorkflow(config CodeAnalysisWorkflowConfig, reporters []reporter.Reporter) *CodeAnalysisWorkflow {
	return &CodeAnalysisWorkflow{
		config: config,
		findings: common.CodeAnalysisFindings{
			SignatureWiseMatchResults: make(map[string][]common.EnrichedSignatureMatchResult),
		},
		reporters: reporters,
	}
}

func (w *CodeAnalysisWorkflow) Execute() (*common.CodeAnalysisFindings, error) {
	err := w.config.Callbacks.OnStart()
	if err != nil {
		w.config.Callbacks.OnErr("failed to execute OnStart callback", err)
		return nil, fmt.Errorf("failed to execute OnStart callback: %w", err)
	}

	err = w.executeInternal()
	if err != nil {
		w.config.Callbacks.OnErr("failed to perform codeanalysis", err)
		return nil, fmt.Errorf("failed to perform codeanalysis: %w", err)
	}

	err = w.reportCodeAnalysisFindings()
	if err != nil {
		w.config.Callbacks.OnErr("failed to report code analysis findings", err)
		return nil, fmt.Errorf("failed to report code analysis findings: %w", err)
	}

	err = w.finishReport()
	if err != nil {
		w.config.Callbacks.OnErr("failed to finish reporting", err)
		return nil, fmt.Errorf("failed to finish reporting: %w", err)
	}

	err = w.config.Callbacks.OnFinish()
	if err != nil {
		w.config.Callbacks.OnErr("failed to execute OnFinish callback", err)
		return nil, fmt.Errorf("failed to execute OnFinish callback: %w", err)
	}

	return &w.findings, nil
}

func (w *CodeAnalysisWorkflow) executeInternal() error {
	fileSystem, err := fs.NewLocalFileSystem(fs.LocalFileSystemConfig{
		AppDirectories: []string{w.config.SourcePath},
	})
	if err != nil {
		return fmt.Errorf("failed to create local filesystem: %w", err)
	}

	allLanguages, err := lang.AllLanguages()
	if err != nil {
		return fmt.Errorf("failed to get all languages: %w", err)
	}

	walker, err := fs.NewSourceWalker(fs.SourceWalkerConfig{}, allLanguages)
	if err != nil {
		return fmt.Errorf("failed to create source walker: %w", err)
	}

	treeWalker, err := parser.NewWalkingParser(walker, allLanguages)
	if err != nil {
		return fmt.Errorf("failed to create tree walker: %w", err)
	}

	callgraphPlugin, err := w.setupCallgraphPlugin()
	if err != nil {
		return fmt.Errorf("failed to setup callgraph plugin: %w", err)
	}

	pluginExecutor, err := plugin.NewTreeWalkPluginExecutor(
		treeWalker,
		[]core.Plugin{callgraphPlugin},
	)
	if err != nil {
		return fmt.Errorf("failed to create plugin executor: %w", err)
	}

	err = pluginExecutor.Execute(context.Background(), fileSystem)
	if err != nil {
		return fmt.Errorf("failed to execute plugin: %w", err)
	}

	return nil
}

func (w *CodeAnalysisWorkflow) setupCallgraphPlugin() (core.Plugin, error) {
	signatureMatcher, err := callgraph.NewSignatureMatcher(w.config.SignaturesToMatch)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature matcher: %w", err)
	}

	var callgraphCallback callgraph.CallgraphCallback = func(_ context.Context, cg *callgraph.CallGraph) error {
		treeData, err := cg.Tree.Data()
		if err != nil {
			return fmt.Errorf("failed to get tree data: %w", err)
		}

		signatureMatches, err := signatureMatcher.MatchSignatures(cg)
		if err != nil {
			return fmt.Errorf("failed to match signatures: %w", err)
		}

		for _, signatureMatch := range signatureMatches {
			w.findings.SignatureWiseMatchResults[signatureMatch.MatchedSignature.Id] = append(w.findings.SignatureWiseMatchResults[signatureMatch.MatchedSignature.Id], common.EnrichedSignatureMatchResult{
				SignatureMatchResult: signatureMatch,
				TreeData:             treeData,
			})
		}

		return nil
	}

	return callgraph.NewCallGraphPlugin(callgraphCallback), nil
}

func (w *CodeAnalysisWorkflow) reportCodeAnalysisFindings() error {
	for _, reporter := range w.reporters {
		err := reporter.RecordCodeAnalysisFindings(&w.findings)
		if err != nil {
			return fmt.Errorf("failed to record code analysis findings in reporter %s: %w", reporter.Name(), err)
		}
	}

	return nil
}

func (w *CodeAnalysisWorkflow) finishReport() error {
	for _, reporter := range w.reporters {
		err := reporter.Finish()
		if err != nil {
			return fmt.Errorf("failed to finish reporter %s: %w", reporter.Name(), err)
		}
	}

	return nil
}
