package models

// ToolCall represents a function call from the AI model
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"` // "function"
	Function FuncCall `json:"function"`
}

// FuncCall represents the function details in a tool call
type FuncCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// ToolResult represents the result of executing a tool
type ToolResult struct {
	ToolCallID string                 `json:"tool_call_id"`
	Name       string                 `json:"name"`
	Content    string                 `json:"content"` // JSON string or text
	Error      string                 `json:"error,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// Tool represents a function definition for the AI model
type Tool struct {
	Type     string   `json:"type"` // "function"
	Function Function `json:"function"`
}

// Function represents a function definition
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"` // JSON schema object
}
