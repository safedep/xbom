package codeanalysis

import (
	"context"
	"errors"
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

// Error constants for code analysis workflow
var (
	ErrOnStartCallback            = errors.New("failed to execute OnStart callback")
	ErrPerformCodeAnalysis        = errors.New("failed to perform codeanalysis")
	ErrReportCodeAnalysisFindings = errors.New("failed to report code analysis findings")
	ErrFinishReporting            = errors.New("failed to finish reporting")
	ErrOnFinishCallback           = errors.New("failed to execute OnFinish callback")
	ErrCreateLocalFileSystem      = errors.New("failed to create local filesystem")
	ErrGetAllLanguages            = errors.New("failed to get all languages")
	ErrCreateSourceWalker         = errors.New("failed to create source walker")
	ErrCreateTreeWalker           = errors.New("failed to create tree walker")
	ErrSetupCallgraphPlugin       = errors.New("failed to setup callgraph plugin")
	ErrCreatePluginExecutor       = errors.New("failed to create plugin executor")
	ErrExecutePlugin              = errors.New("failed to execute plugin")
	ErrCreateSignatureMatcher     = errors.New("failed to create signature matcher")
	ErrGetTreeData                = errors.New("failed to get tree data")
	ErrMatchSignatures            = errors.New("failed to match signatures")
	ErrRecordCodeAnalysisFindings = errors.New("failed to record code analysis findings in reporter")
	ErrFinishReporter             = errors.New("failed to finish reporter")
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
	err := w.config.Callbacks.dispatchOnStart()
	if err != nil {
		w.config.Callbacks.dispatchOnErr(ErrOnStartCallback.Error(), err)
		return nil, fmt.Errorf("%w: %w", ErrOnStartCallback, err)
	}

	err = w.executeInternal()
	if err != nil {
		w.config.Callbacks.dispatchOnErr(ErrPerformCodeAnalysis.Error(), err)
		return nil, fmt.Errorf("%w: %w", ErrPerformCodeAnalysis, err)
	}

	err = w.reportCodeAnalysisFindings()
	if err != nil {
		w.config.Callbacks.dispatchOnErr(ErrReportCodeAnalysisFindings.Error(), err)
		return nil, fmt.Errorf("%w: %w", ErrReportCodeAnalysisFindings, err)
	}

	err = w.finishReport()
	if err != nil {
		w.config.Callbacks.dispatchOnErr(ErrFinishReporting.Error(), err)
		return nil, fmt.Errorf("%w: %w", ErrFinishReporting, err)
	}

	err = w.config.Callbacks.dispatchOnFinish()
	if err != nil {
		w.config.Callbacks.dispatchOnErr(ErrOnFinishCallback.Error(), err)
		return nil, fmt.Errorf("%w: %w", ErrOnFinishCallback, err)
	}

	return &w.findings, nil
}

func (w *CodeAnalysisWorkflow) executeInternal() error {
	fileSystem, err := fs.NewLocalFileSystem(fs.LocalFileSystemConfig{
		AppDirectories: []string{w.config.SourcePath},
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateLocalFileSystem, err)
	}

	allLanguages, err := lang.AllLanguages()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetAllLanguages, err)
	}

	walker, err := fs.NewSourceWalker(fs.SourceWalkerConfig{}, allLanguages)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateSourceWalker, err)
	}

	treeWalker, err := parser.NewWalkingParser(walker, allLanguages)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateTreeWalker, err)
	}

	callgraphPlugin, err := w.setupCallgraphPlugin()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSetupCallgraphPlugin, err)
	}

	pluginExecutor, err := plugin.NewTreeWalkPluginExecutor(
		treeWalker,
		[]core.Plugin{callgraphPlugin},
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreatePluginExecutor, err)
	}

	err = pluginExecutor.Execute(context.Background(), fileSystem)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrExecutePlugin, err)
	}

	return nil
}

func (w *CodeAnalysisWorkflow) setupCallgraphPlugin() (core.Plugin, error) {
	signatureMatcher, err := callgraph.NewSignatureMatcher(w.config.SignaturesToMatch)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSignatureMatcher, err)
	}

	var callgraphCallback callgraph.CallgraphCallback = func(_ context.Context, cg *callgraph.CallGraph) error {
		treeData, err := cg.Tree.Data()
		if err != nil {
			return fmt.Errorf("%w: %w", ErrGetTreeData, err)
		}

		signatureMatches, err := signatureMatcher.MatchSignatures(cg)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrMatchSignatures, err)
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
			return fmt.Errorf("%w %s: %w", ErrRecordCodeAnalysisFindings, reporter.Name(), err)
		}
	}

	return nil
}

func (w *CodeAnalysisWorkflow) finishReport() error {
	for _, reporter := range w.reporters {
		err := reporter.Finish()
		if err != nil {
			return fmt.Errorf("%w %s: %w", ErrFinishReporter, reporter.Name(), err)
		}
	}

	return nil
}
