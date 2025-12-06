#!/bin/bash

# Test Kolosal.ai API directly via curl
# Usage: ./test-kolosal-api.sh YOUR_API_KEY
# Based on working curl test: /v1/chat/completions with model "GLM 4.6"

API_KEY="${1:-${KOLOSAL_API_KEY}}"

if [ -z "$API_KEY" ]; then
    echo "Error: API key required"
    echo "Usage: $0 YOUR_API_KEY"
    echo "   or: KOLOSAL_API_KEY=your_key $0"
    exit 1
fi

echo "Testing Kolosal.ai API..."
echo "API Key: ${API_KEY:0:10}..."
echo ""

# Test 1: /chat/completions endpoint with GLM 4.6 (WORKING CONFIGURATION)
echo "=== Test 1: /chat/completions endpoint with GLM 4.6 (WORKING) ==="
curl -X POST "https://api.kolosal.ai/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "GLM 4.6",
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || cat
echo ""
echo ""

# Test 2: /chat/completions endpoint with different models
echo "=== Test 2: /chat/completions with qwen model ==="
curl -X POST "https://api.kolosal.ai/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "qwen/qwen-3-vl-30b-a3b-instruct",
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || cat
echo ""
echo ""

# Test 3: /chat/completions endpoint with system message
echo "=== Test 3: /chat/completions with system message ==="
curl -X POST "https://api.kolosal.ai/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "GLM 4.6",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello"}
    ]
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || cat
echo ""
echo ""

# Test 4: /chat/completions with max_tokens and temperature
echo "=== Test 4: /chat/completions with max_tokens and temperature ==="
curl -X POST "https://api.kolosal.ai/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "GLM 4.6",
    "messages": [
      {"role": "user", "content": "Hello"}
    ],
    "max_tokens": 2000,
    "temperature": 0.7
  }' \
  -w "\n\nHTTP Status: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || cat
echo ""
echo ""

echo "Done! All tests completed."
echo ""
echo "Expected working configuration:"
echo "  - Endpoint: https://api.kolosal.ai/v1/chat/completions"
echo "  - Model: GLM 4.6"
echo "  - Authorization: Bearer YOUR_API_KEY"
