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
	System    string `json:"system"`
	Prompt    string `json:"prompt"`
	Model     string `json:"model"`
	Streaming bool   `json:"stream"`
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

func (o *Ollama) Complete(system, prompt string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", o.baseURL)
	requestBody := CompletionRequest{
		Prompt: prompt,
		Model:  o.completeModel,
		System: system,
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

	var completionResponse CompletionResponse
	err = json.Unmarshal(body, &completionResponse)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal response: %v ", err)
	}

	return completionResponse.Response, nil
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
