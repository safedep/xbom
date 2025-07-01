package codeanalysis

type CodeAnalysisCallbackRegistry struct {
	OnStart  func() error
	OnFinish func() error
	OnErr    func(msg string, err error)
}

func (c *CodeAnalysisCallbackRegistry) dispatchOnStart() error {
	if c.OnStart != nil {
		return c.OnStart()
	}
	return nil
}

func (c *CodeAnalysisCallbackRegistry) dispatchOnFinish() error {
	if c.OnFinish != nil {
		return c.OnFinish()
	}
	return nil
}

func (c *CodeAnalysisCallbackRegistry) dispatchOnErr(msg string, err error) {
	if c.OnErr != nil {
		c.OnErr(msg, err)
	}
}
