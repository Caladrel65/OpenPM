package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ...existing code...

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	Stream      bool          `json:"stream"`
}

type LLMResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func callLLM(endpoint, model, userPrompt string) (string, string, error) {
	reqBody := ChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a project manager for a small software team. Your job is to check in with the developer, ask for status updates on their current tasks, and offer helpful direction if needed. Always be encouraging and concise."},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
		MaxTokens:   -1,
		Stream:      false,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	// Save raw response for debugging
	rawResponse := string(body)

	// Parse message content
	var llmResp LLMResponse
	err = json.Unmarshal(body, &llmResp)
	if err != nil || len(llmResp.Choices) == 0 {
		return "", rawResponse, fmt.Errorf("could not parse LLM response: %v", err)
	}
	messageContent := llmResp.Choices[0].Message.Content
	return messageContent, rawResponse, nil
}

const (
	llmEndpoint = "http://127.0.0.1:1234/v1/chat/completions"
	modelID     = "gemma-3-12b-it-qat"
)

func main() {
	fmt.Println("Starting!")
	fmt.Println("Calling LLM at:", llmEndpoint)
	userPrompt := "Here is my daily status update: [replace with your actual update]. What should I focus on next?" + "For now, I'm using a placeholder for the update, which will actually include RAG details from the github repo. Please provide a dummy response as though I gave information about e.g. issues, code, commits, etc!"
	message, rawResponse, err := callLLM(llmEndpoint, modelID, userPrompt)
	if err != nil {
		fmt.Println("Error calling LLM:", err)
		fmt.Println("Raw response:", rawResponse)
		os.Exit(1)
	}
	fmt.Println("PM Message:", message)
	// For debugging, you can also print the full response if needed
	// fmt.Println("LLM Raw Response:", rawResponse)
}
