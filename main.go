package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/v63/github"
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



func getProjectStatus() (string, error) {
	content, err := os.ReadFile("PROJECT_STATUS.md")
	if err != nil {
		// If the file doesn't exist, return an empty string
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(content), nil
}

func getGitHubIssues() (string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")

	if token == "" || owner == "" || repo == "" {
		return "(GitHub issues not configured)", nil
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	issues, _, err := client.Issues.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{
		State: "open",
	})
	if err != nil {
		return "", err
	}

	if len(issues) == 0 {
		return "No open issues found.", nil
	}

	var issueList string
	for _, issue := range issues {
		issueList += fmt.Sprintf("- #%d: %s\n", issue.GetNumber(), issue.GetTitle())
	}

	return issueList, nil
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

	

	projectStatus, err := getProjectStatus()
	if err != nil {
		fmt.Println("Error fetching project status:", err)
		os.Exit(1)
	}

	issues, err := getGitHubIssues()
	if err != nil {
		fmt.Println("Error fetching GitHub issues:", err)
		os.Exit(1)
	}

	userMessage := ChatMessage{
		Role:    "user",
		Content: fmt.Sprintf(`Today is %s. Here is the current project status:

%s

Here are the open GitHub issues:

%s`, time.Now().Format("Monday, January 2, 2006"), projectStatus, issues),
	}

	messages := []ChatMessage{
		{Role: "system", Content: "You are a friendly project manager AI. You are talking directly to a solo hobbyist developer. Your role is to check in on their progress, understand that they have limited time, and provide encouragement. Address the developer directly. Keep your messages concise, supportive, and focused on helping them take the next small step. If the user's message doesn't mention GitHub issues, check the conversation history. If GitHub integration hasn't been discussed recently, gently suggest connecting to GitHub for better project management."},
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
	discordMessage := fmt.Sprintf("<@119165799284867072> %s", pmResponse)
	fmt.Println("Sending message to Discord...")
	if err := sendToDiscord(webhookURL, discordMessage); err != nil {
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
