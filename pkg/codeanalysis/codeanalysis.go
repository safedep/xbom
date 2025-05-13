package codeanalysis

import (
	"context"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/safedep/code/core"
	"github.com/safedep/code/fs"
	"github.com/safedep/code/lang"
	"github.com/safedep/code/parser"
	"github.com/safedep/code/plugin"
	"github.com/safedep/code/plugin/callgraph"
)

type CodeAnalysisWorkflow struct {
	config   CodeAnalysisWorkflowConfig
	findings CodeAnalysisFindings
}

func NewCodeAnalysisWorkflow(config CodeAnalysisWorkflowConfig) *CodeAnalysisWorkflow {
	return &CodeAnalysisWorkflow{
		config: config,
		findings: CodeAnalysisFindings{
			SignatureWiseMatchResults: make(map[string][]callgraph.SignatureMatchResult),
		},
	}
}

func (w *CodeAnalysisWorkflow) Execute() error {
	err := w.config.Callbacks.OnStart()
	if err != nil {
		return fmt.Errorf("failed to execute OnStart callback: %w", err)
	}

	err = w.executeInternal()
	if err != nil {
		w.config.Callbacks.OnErr("failed to perform codeanalysis", err)
		return fmt.Errorf("failed to perform codeanalysis: %w", err)
	}

	err = w.config.Callbacks.OnFinish()
	if err != nil {
		return fmt.Errorf("failed to execute OnFinish callback: %w", err)
	}

	return nil
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

	callgraphPlugin, err := w.SetupCallgraphPlugin()
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

func (w *CodeAnalysisWorkflow) SetupCallgraphPlugin() (core.Plugin, error) {
	signatureMatcher, err := callgraph.NewSignatureMatcher(w.config.SignaturesToMatch)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature matcher: %w", err)
	}

	var callgraphCallback callgraph.CallgraphCallback = func(_ context.Context, cg *callgraph.CallGraph) error {
		signatureMatches, err := signatureMatcher.MatchSignatures(cg)
		if err != nil {
			return fmt.Errorf("failed to match signatures: %w", err)
		}

		for _, signatureMatch := range signatureMatches {
			w.findings.SignatureWiseMatchResults[signatureMatch.MatchedSignature.Id] = append(w.findings.SignatureWiseMatchResults[signatureMatch.MatchedSignature.Id], signatureMatch)
		}

		return nil
	}

	return callgraph.NewCallGraphPlugin(callgraphCallback), nil
}

func (w *CodeAnalysisWorkflow) Finish(summarize bool) *CodeAnalysisFindings {
	if summarize {
		w.summarize()
	}
	return &w.findings
}

func (w *CodeAnalysisWorkflow) summarize() {
	// Create a table writer for signatures
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

	// Process the match data
	for _, signatureResults := range w.findings.SignatureWiseMatchResults {
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
						evidence.Namespace,
						evidenceDetailString,
					})
					sigTable.AppendSeparator()
				}
			}
		}
	}

	sigTable.Render()
}
