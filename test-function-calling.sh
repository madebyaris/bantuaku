#!/bin/bash

# Test Function Calling Support for Kolosal.ai and OpenRouter
# Usage: ./test-function-calling.sh

echo "🧪 Function Calling Test Script"
echo "================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    exit 1
fi

# Check environment variables
if [ -z "$KOLOSAL_API_KEY" ] && [ -z "$OPENROUTER_API_KEY" ]; then
    echo "⚠️  Warning: Neither KOLOSAL_API_KEY nor OPENROUTER_API_KEY is set"
    echo "   Set at least one to test function calling"
    echo ""
fi

if [ -n "$KOLOSAL_API_KEY" ]; then
    echo "✅ KOLOSAL_API_KEY is set"
else
    echo "⚠️  KOLOSAL_API_KEY is not set (will skip Kolosal test)"
fi

if [ -n "$OPENROUTER_API_KEY" ]; then
    echo "✅ OPENROUTER_API_KEY is set"
else
    echo "⚠️  OPENROUTER_API_KEY is not set (will skip OpenRouter test)"
fi

echo ""
echo "Running tests..."
echo ""

# Run the test from project root
go run backend/testing/test_function_calling.go

echo ""
echo "📝 Test Results Summary:"
echo "   - Check output above for function calling support"
echo "   - Look for '✅ FUNCTION CALLING SUPPORTED!' messages"
echo "   - Tool calls will show function names and arguments"
