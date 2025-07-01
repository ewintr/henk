package agent

import (
	"bytes"
	"fmt"
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
