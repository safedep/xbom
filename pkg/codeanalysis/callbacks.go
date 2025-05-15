package codeanalysis

type CodeAnalysisCallbackRegistry struct {
	OnStart  func() error
	OnFinish func() error
	OnErr    func(msg string, err error) error
}
