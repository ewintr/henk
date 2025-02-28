package parse

import (
	"encoding/json"
	"fmt"

	"go-mod.ewintr.nl/henk/llm"
)

const (
	system = "You are an expert in software development and can understand and summarize any file one might encounter in a software project"
	schema = `{
  "type": "object",
  "properties": {
    "sentenceSummary": {
      "type": "string"
    },
    "paragraphSummary": {
      "type": "string"
    }
  }
}`
	promptFmtSource = `The following is the content of a file with Go source code. Give two summaries for it:
- One simple sentence that describes the source code functionality in the file
- One paragraph that also the source code functionality, but that goes a little more in depth

Don't explain that it is source code, or that it is Go. Focus on conveying the functionality that the code implements.

This is the file %s:

---

%s

---

Respond in JSON.`
)

type DescribeResponse struct {
	Sentence  string `json:"sentenceSummary"`
	Paragraph string `json:"paragraphSummary"`
}

func Describe(file *File, client *llm.Ollama) (string, string, error) {
	prompt := fmt.Sprintf(promptFmtSource, file.Name(), file.Content)
	res, err := client.Complete(system, prompt, []byte(schema))
	if err != nil {
		return "", "", fmt.Errorf("could not complete request: %v", err)
	}
	var sum DescribeResponse
	if err := json.Unmarshal([]byte(res), &sum); err != nil {
		return "", "", fmt.Errorf("could not unmarshal response: %v", err)
	}
	return sum.Sentence, sum.Paragraph, nil
}
