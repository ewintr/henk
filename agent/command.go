package agent

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"go-mod.ewintr.nl/henk/agent/llm"
)

var (
	listModelsTpl *template.Template
	helpTpl       *template.Template
)

func init() {
	listModelsTpl = template.Must(template.New("listModels").Parse(`Available models:

{{ range . }}
- {{ .Provider }}: {{ .Model }}{{ if .Short }} ({{ .Short }}){{ end }}
{{ end }}
`))

	helpTpl = template.Must(template.New("help").Parse(`Available commands:   

{{ range $key, $value := . }}
- **{{ $key }}**: {{ $value }}
{{ end }}

Also, press ctrl-e to open an editor to edit your message.
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
	case "help":
		a.showHelp()
	case "models":
		a.listModels()
	case "switch":
		a.switchModel(args)
	case "clear":
		a.clearContext()
	case "copy":
		a.copyLastMessage()
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

func (a *Agent) showHelp() {
	cmds := map[string]string{
		"/help":                      "Show this help message",
		"/status":                    "Show current LLM",
		"/models":                    "List available models",
		"/switch [model]":            "Switch to model with complete name  or short name",
		"/switch [provider] [model]": "Switch to specific provider model",
		"/clear":                     "Reset conversation, clear the context",
		"/copy":                      "Copy last message to the clipboard",
		"/quit":                      "Exit the agent",
	}
	msg := bytes.NewBuffer([]byte{})
	if err := helpTpl.Execute(msg, cmds); err != nil {
		a.out <- Message{Type: TypeError, Body: fmt.Sprintf("could not execute help template: %v", err.Error())}
		return
	}
	a.out <- Message{Type: TypeGeneral, Body: msg.String()}
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

func (a *Agent) clearContext() {
	a.conversation = make([]llm.Message, 0)
	a.displayGen("Context cleared")
}

func (a *Agent) copyLastMessage() {
	if a.config.ClipboardCommand == "" {
		a.displayError("No clipboard command configured in config file")
		return
	}

	// Find the last assistant message with text content
	var lastText string
	for i := len(a.conversation) - 1; i >= 0; i-- {
		msg := a.conversation[i]
		if msg.Role == llm.RoleAssistant {
			for _, content := range msg.Content {
				if content.Type == llm.ContentTypeText && content.Text != "" {
					lastText = content.Text
					break
				}
			}
			if lastText != "" {
				break
			}
		}
	}

	if lastText == "" {
		a.displayError("No assistant message found to copy")
		return
	}

	// Execute the clipboard command with text piped to stdin
	cmd := exec.Command("sh", "-c", a.config.ClipboardCommand)
	cmd.Stdin = strings.NewReader(lastText)
	if err := cmd.Run(); err != nil {
		a.displayError(fmt.Sprintf("Failed to copy to clipboard: %v", err))
		return
	}

	a.displayGen("Last message copied to clipboard")
}
