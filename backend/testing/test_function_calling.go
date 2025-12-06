//go:build ignore
// +build ignore

package main

// This is a standalone test script to verify function calling support
// Run with: go run backend/testing/test_function_calling.go
// Or use: ./test-function-calling.sh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Test function calling capabilities for Kolosal.ai and OpenRouter

type ChatCompletionRequest struct {
	Model       string                  `json:"model"`
	Messages    []ChatCompletionMessage `json:"messages"`
	Tools       []Tool                  `json:"tools,omitempty"`
	ToolChoice  interface{}             `json:"tool_choice,omitempty"` // Can be "none", "auto", "required", or object
	MaxTokens   int                     `json:"max_tokens,omitempty"`
	Temperature float64                 `json:"temperature,omitempty"`
}

type ChatCompletionMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function FuncCall `json:"function"`
}

type FuncCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatCompletionResponse struct {
	ID      string       `json:"id,omitempty"`
	Choices []ChatChoice `json:"choices"`
	Usage   *Usage       `json:"usage,omitempty"`
}

type ChatChoice struct {
	Index        int                   `json:"index,omitempty"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

func main() {
	fmt.Println("🧪 Testing Function Calling Support")
	fmt.Println("=====================================\n")

	// Test tools definition
	tools := []Tool{
		{
			Type: "function",
			Function: Function{
				Name:        "update_company_info",
				Description: "Update company information fields (industry, location, business model)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"industry": map[string]interface{}{
							"type":        "string",
							"description": "Business industry/type (e.g., 'Kuliner', 'Retail', 'Jasa')",
						},
						"city": map[string]interface{}{
							"type":        "string",
							"description": "City where business operates",
						},
						"location_region": map[string]interface{}{
							"type":        "string",
							"description": "Region/province",
						},
					},
					"required": []string{"industry", "city"},
				},
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "create_product",
				Description: "Create a new product or service for the company",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Product/service name",
						},
						"category": map[string]interface{}{
							"type":        "string",
							"description": "Product category",
						},
						"unit_price": map[string]interface{}{
							"type":        "number",
							"description": "Price per unit",
						},
					},
					"required": []string{"name"},
				},
			},
		},
	}

	testMessage := "Saya punya bisnis kuliner di Jakarta. Produk utama saya adalah Nasi Goreng Spesial dengan harga 25000."

	// Test Kolosal.ai
	fmt.Println("📡 Testing Kolosal.ai API (GLM 4.6)")
	fmt.Println("-----------------------------------")
	kolosalAPIKey := os.Getenv("KOLOSAL_API_KEY")
	if kolosalAPIKey == "" {
		fmt.Println("❌ KOLOSAL_API_KEY not set, skipping Kolosal test")
	} else {
		testKolosal(kolosalAPIKey, tools, testMessage)
	}

	fmt.Println("\n")

	// Test OpenRouter with x-ai/grok-4-fast
	fmt.Println("📡 Testing OpenRouter API (x-ai/grok-4-fast)")
	fmt.Println("--------------------------------------------")
	openRouterAPIKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterAPIKey == "" {
		fmt.Println("❌ OPENROUTER_API_KEY not set, skipping OpenRouter test")
	} else {
		testOpenRouter(openRouterAPIKey, tools, testMessage)
	}

	fmt.Println("\n✅ Testing complete!")
}

func testKolosal(apiKey string, tools []Tool, userMessage string) {
	url := "https://api.kolosal.ai/v1/chat/completions"

	req := ChatCompletionRequest{
		Model: "GLM 4.6",
		Messages: []ChatCompletionMessage{
			{
				Role:    "system",
				Content: "Kamu adalah Asisten Bantuaku. Gunakan tools yang tersedia untuk menyimpan informasi perusahaan dan produk.",
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		Tools:       tools,
		ToolChoice:  "auto", // Let model decide when to use tools
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("❌ Failed to marshal request: %v\n", err)
		return
	}

	fmt.Printf("📤 Request URL: %s\n", url)
	fmt.Printf("📤 Model: %s\n", req.Model)
	fmt.Printf("📤 Tool Choice: %v\n", req.ToolChoice)
	fmt.Printf("📤 Tools Count: %d\n", len(req.Tools))
	fmt.Printf("📤 Request Body (first 500 chars): %s...\n\n", string(reqBody)[:min(500, len(string(reqBody)))])

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("📥 Response Status: %d %s\n", resp.StatusCode, resp.Status)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ API Error Response:\n%s\n", string(bodyBytes))
		return
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(bodyBytes))
		return
	}

	if len(chatResp.Choices) == 0 {
		fmt.Printf("❌ No choices in response\n")
		fmt.Printf("Raw response: %s\n", string(bodyBytes))
		return
	}

	choice := chatResp.Choices[0]
	fmt.Printf("📥 Finish Reason: %s\n", choice.FinishReason)
	fmt.Printf("📥 Message Role: %s\n", choice.Message.Role)
	fmt.Printf("📥 Message Content: %s\n", choice.Message.Content)

	if len(choice.Message.ToolCalls) > 0 {
		fmt.Printf("✅ FUNCTION CALLING SUPPORTED!\n")
		fmt.Printf("📥 Tool Calls Count: %d\n", len(choice.Message.ToolCalls))
		for i, toolCall := range choice.Message.ToolCalls {
			fmt.Printf("\n  Tool Call #%d:\n", i+1)
			fmt.Printf("    ID: %s\n", toolCall.ID)
			fmt.Printf("    Type: %s\n", toolCall.Type)
			fmt.Printf("    Function Name: %s\n", toolCall.Function.Name)
			fmt.Printf("    Function Arguments: %s\n", toolCall.Function.Arguments)
		}
	} else {
		fmt.Printf("⚠️  No tool calls in response\n")
		fmt.Printf("   This could mean:\n")
		fmt.Printf("   1. Function calling not supported by Kolosal.ai API\n")
		fmt.Printf("   2. Model chose not to use tools (check finish_reason)\n")
		fmt.Printf("   3. API ignored tools parameter\n")
	}

	if chatResp.Usage != nil {
		fmt.Printf("\n📊 Token Usage:\n")
		fmt.Printf("   Prompt: %d\n", chatResp.Usage.PromptTokens)
		fmt.Printf("   Completion: %d\n", chatResp.Usage.CompletionTokens)
		fmt.Printf("   Total: %d\n", chatResp.Usage.TotalTokens)
	}
}

func testOpenRouter(apiKey string, tools []Tool, userMessage string) {
	url := "https://openrouter.ai/api/v1/chat/completions"

	req := ChatCompletionRequest{
		Model: "x-ai/grok-4-fast",
		Messages: []ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are Bantuaku Assistant. Use available tools to store company and product information.",
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		Tools:       tools,
		ToolChoice:  "auto",
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("❌ Failed to marshal request: %v\n", err)
		return
	}

	fmt.Printf("📤 Request URL: %s\n", url)
	fmt.Printf("📤 Model: %s\n", req.Model)
	fmt.Printf("📤 Tool Choice: %v\n", req.ToolChoice)
	fmt.Printf("📤 Tools Count: %d\n", len(req.Tools))
	fmt.Printf("📤 Request Body (first 500 chars): %s...\n\n", string(reqBody)[:min(500, len(string(reqBody)))])

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	httpReq.Header.Set("HTTP-Referer", "https://bantuaku.com") // Required by OpenRouter
	httpReq.Header.Set("X-Title", "Bantuaku Function Calling Test")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("📥 Response Status: %d %s\n", resp.StatusCode, resp.Status)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ API Error Response:\n%s\n", string(bodyBytes))
		return
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(bodyBytes))
		return
	}

	if len(chatResp.Choices) == 0 {
		fmt.Printf("❌ No choices in response\n")
		fmt.Printf("Raw response: %s\n", string(bodyBytes))
		return
	}

	choice := chatResp.Choices[0]
	fmt.Printf("📥 Finish Reason: %s\n", choice.FinishReason)
	fmt.Printf("📥 Message Role: %s\n", choice.Message.Role)
	fmt.Printf("📥 Message Content: %s\n", choice.Message.Content)

	if len(choice.Message.ToolCalls) > 0 {
		fmt.Printf("✅ FUNCTION CALLING SUPPORTED!\n")
		fmt.Printf("📥 Tool Calls Count: %d\n", len(choice.Message.ToolCalls))
		for i, toolCall := range choice.Message.ToolCalls {
			fmt.Printf("\n  Tool Call #%d:\n", i+1)
			fmt.Printf("    ID: %s\n", toolCall.ID)
			fmt.Printf("    Type: %s\n", toolCall.Type)
			fmt.Printf("    Function Name: %s\n", toolCall.Function.Name)
			fmt.Printf("    Function Arguments: %s\n", toolCall.Function.Arguments)
		}
	} else {
		fmt.Printf("⚠️  No tool calls in response\n")
		fmt.Printf("   This could mean:\n")
		fmt.Printf("   1. Function calling not supported by model\n")
		fmt.Printf("   2. Model chose not to use tools (check finish_reason)\n")
		fmt.Printf("   3. API ignored tools parameter\n")
	}

	if chatResp.Usage != nil {
		fmt.Printf("\n📊 Token Usage:\n")
		fmt.Printf("   Prompt: %d\n", chatResp.Usage.PromptTokens)
		fmt.Printf("   Completion: %d\n", chatResp.Usage.CompletionTokens)
		fmt.Printf("   Total: %d\n", chatResp.Usage.TotalTokens)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
