package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

const (
	llmEndpoint = "http://127.0.0.1:1234/v1/chat/completions"
	modelID     = "gemma-3-12b-it-qat"
	historyFile = "conversation_history.json"
	maxHistory  = 10 // Keep the last 10 messages (5 pairs)
)

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

type DiscordWebhook struct {
	Content string `json:"content"`
}

func callLLM(endpoint, model string, messages []ChatMessage) (string, string, error) {
	reqBody := ChatRequest{
		Model:       model,
		Messages:    messages,
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
	rawResponse := string(body)

	var llmResp LLMResponse
	err = json.Unmarshal(body, &llmResp)
	if err != nil || len(llmResp.Choices) == 0 {
		return "", rawResponse, fmt.Errorf("could not parse LLM response: %v", err)
	}
	messageContent := llmResp.Choices[0].Message.Content
	return messageContent, rawResponse, nil
}

func getLatestCommits() (string, error) {
	cmd := exec.Command("git", "log", "-n", "5", "--pretty=format:%h - %an, %ar : %s")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func loadHistory() ([]ChatMessage, error) {
	if _, err := os.Stat(historyFile);
		os.IsNotExist(err) {
		return []ChatMessage{}, nil
	}
	file, err := os.ReadFile(historyFile)
	if err != nil {
		return nil, err
	}
	var history []ChatMessage
	err = json.Unmarshal(file, &history)
	return history, err
}

func saveHistory(history []ChatMessage) error {
	if len(history) > maxHistory {
		history = history[len(history)-maxHistory:]
	}
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(historyFile, data, 0644)
}

func sendToDiscord(webhookURL, message string) error {
	data, err := json.Marshal(DiscordWebhook{Content: message})
	if err != nil {
		return err
	}
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord webhook failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func main() {
	fmt.Println("Starting!")

	history, err := loadHistory()
	if err != nil {
		fmt.Println("Error loading history:", err)
		os.Exit(1)
	}

	fmt.Println("Fetching latest commits...")
	commits, err := getLatestCommits()
	if err != nil {
		fmt.Println("Error fetching commits:", err)
		os.Exit(1)
	}

	userMessage := ChatMessage{
		Role:    "user",
		Content: fmt.Sprintf("Here are the 5 latest commits for the project:\n\n%s\n\nBased on these, and our previous conversation, what should I focus on next?", commits),
	}

	messages := []ChatMessage{
		{Role: "system", Content: "You are a project manager for a small software team. Your job is to check in with the developer, ask for status updates on their current tasks, and offer helpful direction if needed. Always be encouraging and concise."},
	}
	messages = append(messages, history...)
	messages = append(messages, userMessage)

	fmt.Println("Calling LLM at:", llmEndpoint)
	pmResponse, rawResponse, err := callLLM(llmEndpoint, modelID, messages)
	if err != nil {
		fmt.Println("Error calling LLM:", err)
		fmt.Println("Raw response:", rawResponse)
		os.Exit(1)
	}

	fmt.Println("PM Message:", pmResponse)

	webhookURL := "https://discord.com/api/webhooks/1395642699738255391/eHv1tXIhl7mw4MSaWuniqwi3UC7o6uRxltfcGbB89Au3nPog-HuGmBrInAyLLUjwLtaB"
	fmt.Println("Sending message to Discord...")
	if err := sendToDiscord(webhookURL, pmResponse); err != nil {
		fmt.Println("Error sending to Discord:", err)
		// Do not exit here, we still want to save the history
	}

	history = append(history, userMessage, ChatMessage{Role: "assistant", Content: pmResponse})
	if err := saveHistory(history); err != nil {
		fmt.Println("Error saving history:", err)
		os.Exit(1)
	}

	fmt.Println("Conversation history saved.")
}
