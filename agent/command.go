package agent

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var (
	listModelsTpl *template.Template
)

func init() {
	listModelsTpl = template.Must(template.New("listModels").Parse(`Available models:

{{ range . }}
- {{ .Provider }}: {{ .Model }}
{{ end }}
`))

}

func (a *Agent) runCommand(input string) {
	cmd, _, _ := strings.Cut(input, " ")
	cmd = strings.TrimPrefix(cmd, "/")
	switch cmd {
	case "quit":
		a.done = true
		a.out <- Message{Type: TypeExit}
	case "models":
		a.listModels()
	case "status":
		a.showStatus()
	}
}

func (a *Agent) listModels() {
	type item struct {
		Provider string
		Model    string
	}
	data := make([]item, 0)
	for _, p := range a.config.Providers {
		for _, m := range p.Models {
			mName := m.Name
			if m.ShortName != "" {
				mName = m.ShortName
			}
			data = append(data, item{
				Provider: p.Name,
				Model:    mName,
			})
		}
	}
	msg := bytes.NewBuffer([]byte{})
	if err := listModelsTpl.Execute(msg, data); err != nil {
		a.out <- Message{Type: TypeError, Body: fmt.Sprintf("could not execute listModels template: %v", err.Error())}
	}

	a.out <- Message{Type: TypeGeneral, Body: msg.String()}
}

func (a *Agent) showStatus() {
	prov, mod, short := a.llmClient.ModelInfo()
	status := fmt.Sprintf("Current LLM: %s - %s", prov, mod)
	if short != "" {
		status = fmt.Sprintf("%s (%s)", status, short)
	}
	a.out <- Message{Type: TypeGeneral, Body: status}
}

func (a *Agent) switchModel(args string) {
}
