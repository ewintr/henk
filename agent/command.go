package agent

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"go-mod.ewintr.nl/henk/agent/llm"
)

var (
	listModelsTpl *template.Template
)

func init() {
	listModelsTpl = template.Must(template.New("listModels").Parse(`Available models:

{{ range . }}
- {{ .Provider }}: {{ .Model }}{{ if .Short }} ({{ .Short }}){{ end }}
{{ end }}
`))

}

func (a *Agent) runCommand(input string) {
	cmd, args, _ := strings.Cut(input, " ")
	cmd = strings.TrimPrefix(cmd, "/")
	switch cmd {
	case "quit":
		a.done = true
		a.out <- Message{Type: TypeExit}
	case "status":
		a.showStatus()
	case "models":
		a.listModels()
	case "switch":
		a.switchModel(args)
	}
}

func (a *Agent) showStatus() {
	prov, mod, short := a.llmClient.ModelInfo()
	status := fmt.Sprintf("Current LLM: %s:  %s", prov, mod)
	if short != "" {
		status = fmt.Sprintf("%s (%s)", status, short)
	}
	a.out <- Message{Type: TypeGeneral, Body: status}
}

func (a *Agent) listModels() {
	type item struct {
		Provider string
		Model    string
		Short    string
	}
	data := make([]item, 0)
	for _, p := range a.config.Providers {
		for _, m := range p.Models {
			data = append(data, item{
				Provider: p.Name,
				Model:    m.Name,
				Short:    m.ShortName,
			})
		}
	}
	msg := bytes.NewBuffer([]byte{})
	if err := listModelsTpl.Execute(msg, data); err != nil {
		a.out <- Message{Type: TypeError, Body: fmt.Sprintf("could not execute listModels template: %v", err.Error())}
	}

	a.out <- Message{Type: TypeGeneral, Body: msg.String()}
}

func (a *Agent) switchModel(args string) {
	args = strings.TrimSpace(args)
	if args == "" {
		a.displayError("Usage: /switch <model> or /switch <provider> <model>")
		return
	}

	var provider llm.Provider
	var modelName string
	var ok bool
	parts := strings.SplitN(args, " ", 2)
	if len(parts) == 2 {
		provider, ok = a.config.Provider(parts[0])
		if !ok {
			a.displayError(fmt.Sprintf("could not find provider %q", parts[0]))
			return
		}
		modelName = parts[1]
	} else {
		modelName = parts[0]
		provider, ok = a.config.ProviderByModelName(modelName)
		if !ok {
			a.displayError(fmt.Sprintf("Could not find provider for model %q", modelName))
			return
		}
	}

	newClient, err := llm.NewLLM(provider, modelName, a.config.SystemPrompt)
	if err != nil {
		a.displayError(fmt.Sprintf("Failed to switch: %q", err.Error()))
	}
	a.llmClient = newClient

	a.showStatus()
}
