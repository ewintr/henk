package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CompletionRequest struct {
	System    string          `json:"system"`
	Prompt    string          `json:"prompt"`
	Model     string          `json:"model"`
	Streaming bool            `json:"stream"`
	Format    json.RawMessage `json:"format,omitempty"`
}

type CompletionResponse struct {
	Response string `json:"response"`
}

type Ollama struct {
	baseURL       string
	embedModel    string
	completeModel string
	client        *http.Client
}

func NewOllama(baseURL, embedModel, completeModel string) *Ollama {
	return &Ollama{
		baseURL:       baseURL,
		embedModel:    embedModel,
		completeModel: completeModel,
		client:        &http.Client{Timeout: 600 * time.Second},
	}
}

func (o *Ollama) Complete(system, prompt string, format json.RawMessage) (string, error) {
	url := fmt.Sprintf("%s/api/generate", o.baseURL)
	requestBody := CompletionRequest{
		Prompt: prompt,
		Model:  o.completeModel,
		System: system,
	}
	if format != nil {
		requestBody.Format = format
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("could not marshal request to json: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("could not post request to ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("received non-successful status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response: %v", err)
	}
	// fmt.Println(string(body))

	var completionResponse CompletionResponse
	err = json.Unmarshal(body, &completionResponse)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal response: %v ", err)
	}

	return completionResponse.Response, nil
}

const snippetSchema = `{
  "type": "object",
  "properties": {
    "snippets": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "identifier": {
            "type": "string"
          },
          "kind": {
            "type": "string",
            "enum": ["function", "type", "constant", "variable", "other"]
          },
          "lineRange": {
            "type": "object",
            "properties": {
              "start": {
                "type": "integer",
                "minimum": 1
              },
              "end": {
                "type": "integer",
                "minimum": 1
              }
            },
            "required": ["start", "end"],
            "additionalProperties": false
          }
        },
        "required": ["identifier", "kind", "lineRange"],
        "additionalProperties": false
      }
    }
  },
  "required": ["snippets"],
  "additionalProperties": false
}`

type Snippet struct {
	Identifier string `json:"identifier"`
	Kind       string `json:"kind"`
	LineRange  struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"lineRange"`
}

type SnippetCompletionResponse struct {
	Snippets []Snippet `json:"snippets"`
}

func (o *Ollama) CompleteWithSnippets(system, prompt string) ([]Snippet, error) {
	resp, err := o.Complete(system, prompt, []byte(snippetSchema))
	if err != nil {
		return nil, err
	}
	var snippetCompletionResponse SnippetCompletionResponse
	err = json.Unmarshal([]byte(resp), &snippetCompletionResponse)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response: %v ", err)
	}

	return snippetCompletionResponse.Snippets, nil
}

func (o *Ollama) Embed(inputText string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model": "text-embedding-3-small",
		"input": inputText,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		o.baseURL+"/v1/embeddings",
		strings.NewReader(string(jsonBody)),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, resp: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(result.Data) == 0 || len(result.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return result.Data[0].Embedding, nil
}
