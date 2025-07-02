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
	provider := a.config.Provider()
	model := provider.DefaultModel()
	modelName := model.Name
	if model.ShortName != "" {
		modelName = model.ShortName
	}

	status := fmt.Sprintf("Current LLM: %s - %s", provider.Name,
		modelName)
	a.out <- Message{Type: TypeGeneral, Body: status}
}

func (a *Agent) switchModel(args string) {
	if args == "" {
		a.out <- Message{Type: TypeError, Body: "Usage: /switch <provider>:<model> or /switch <short_name>"}
		return
	}

	// Try to find model by short name first
	var targetProvider *llm.Provider
	var targetModel *llm.Model

	for i := range a.config.Providers {
		for j := range a.config.Providers[i].Models {
			model := &a.config.Providers[i].Models[j]
			if model.ShortName == args || model.Name == args {
				targetProvider = &a.config.Providers[i]
				targetModel = model
				break
			}
		}
		if targetProvider != nil {
			break
		}
	}

	// Try provider:model format
	if targetProvider == nil && strings.Contains(args, ":") {
		parts := strings.SplitN(args, ":", 2)
		providerName, modelName := parts[0], parts[1]

		for i := range a.config.Providers {
			if a.config.Providers[i].Name == providerName {
				for j := range a.config.Providers[i].Models {
					model := &a.config.Providers[i].Models[j]
					if model.Name == modelName || model.ShortName ==
						modelName {
						targetProvider = &a.config.Providers[i]
						targetModel = model
						break
					}
				}
				break
			}
		}
	}

	if targetProvider == nil {
		a.out <- Message{Type: TypeError, Body: fmt.Sprintf("Model '%s' not found", args)}
		return
	}

	// Reinitialize LLM client
	newLLMClient, err := llm.NewLLM(*targetProvider, a.config.
		SystemPrompt)
	if err != nil {
		a.out <- Message{Type: TypeError, Body: fmt.Sprintf("Failed to initialize LLM: %v", err)}
		return
	}

	a.llmClient = newLLMClient

	modelDisplay := targetModel.Name
	if targetModel.ShortName != "" {
		modelDisplay = targetModel.ShortName
	}

	a.out <- Message{Type: TypeGeneral, Body: fmt.Sprintf("Switched to: %s - %s", targetProvider.Name, modelDisplay)}
}
