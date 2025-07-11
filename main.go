package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ...existing code...

func callLLM(endpoint string) (string, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func main() {
	llmEndpoint := os.Getenv("LLM_ENDPOINT")
	schedule := os.Getenv("SCHEDULE")
	if llmEndpoint == "" || schedule == "" {
		fmt.Println("LLM_ENDPOINT and SCHEDULE environment variables must be set.")
		os.Exit(1)
	}

	interval, err := time.ParseDuration(schedule)
	if err != nil {
		fmt.Println("Invalid schedule format:", err)
		os.Exit(1)
	}

	for {
		fmt.Println("Calling LLM at:", llmEndpoint)
		response, err := callLLM(llmEndpoint)
		if err != nil {
			fmt.Println("Error calling LLM:", err)
		} else {
			fmt.Println("LLM Response:", response)
		}
		fmt.Printf("Waiting %s until next call...\n", interval)
		time.Sleep(interval)
	}
}
