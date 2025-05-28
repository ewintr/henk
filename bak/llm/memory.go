package llm

// Memory is a mock implementation of EmbedderCompleter
type Memory struct {
	EmbedReturns    [][]float32
	EmbedError      error
	CompleteReturns []string
	CompleteError   error
}

func (m *Memory) Embed(input string) ([]float32, error) {
	if m.EmbedError != nil {
		return nil, m.EmbedError
	}
	res := m.EmbedReturns[0]
	if len(m.EmbedReturns) > 1 {
		m.EmbedReturns = m.EmbedReturns[1:]
	}
	return res, nil
}

func (m *Memory) Complete(input string) (string, error) {
	if m.CompleteError != nil {
		return "", m.CompleteError
	}
	res := m.CompleteReturns[0]
	if len(m.CompleteReturns) > 1 {
		m.CompleteReturns = m.CompleteReturns[1:]
	}
	return res, nil
}
