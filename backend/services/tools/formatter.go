package tools

import (
	"encoding/json"

	"github.com/bantuaku/backend/models"
	"github.com/bantuaku/backend/services/chat"
)

// FormatToolResult formats a tool execution result as a tool message for the AI
func FormatToolResult(result *models.ToolResult) chat.ChatCompletionMessage {
	// If there's an error, include it in the content
	content := result.Content
	if result.Error != "" {
		errorData := map[string]interface{}{
			"error":   result.Error,
			"tool":    result.Name,
			"message": "Tool execution failed",
		}
		errorJSON, _ := json.Marshal(errorData)
		content = string(errorJSON)
	}

	return chat.ChatCompletionMessage{
		Role:       "tool",
		Content:    content,
		ToolCallID: result.ToolCallID, // OpenAI format requires tool_call_id in tool messages
	}
}

// FormatToolResults formats multiple tool results as tool messages
func FormatToolResults(results []*models.ToolResult) []chat.ChatCompletionMessage {
	messages := make([]chat.ChatCompletionMessage, len(results))
	for i, result := range results {
		messages[i] = FormatToolResult(result)
	}
	return messages
}
