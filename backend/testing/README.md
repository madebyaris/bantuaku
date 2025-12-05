# Function Calling Test

This directory contains test scripts to verify function calling support for:
- **Kolosal.ai** (GLM 4.6 model)
- **OpenRouter** (x-ai/grok-4-fast model)

## Prerequisites

1. **Go** installed (for Go test script)
2. **Python 3** installed (for Python test script - optional)
3. API keys set as environment variables:
   - `KOLOSAL_API_KEY` - Your Kolosal.ai API key
   - `OPENROUTER_API_KEY` - Your OpenRouter API key

## Running the Test

### Option 1: Go Script (Recommended)

```bash
# From project root
./test-function-calling.sh

# Or directly
go run backend/testing/test_function_calling.go
```

### Option 2: Python Script

```bash
# From project root
python3 backend/testing/function_calling_test.py
```

## What the Test Does

1. **Defines Test Tools**:
   - `update_company_info` - Updates company information
   - `create_product` - Creates a new product

2. **Sends Test Message**: 
   - "Saya punya bisnis kuliner di Jakarta. Produk utama saya adalah Nasi Goreng Spesial dengan harga 25000."
   - (Translation: "I have a culinary business in Jakarta. My main product is Special Fried Rice priced at 25000.")

3. **Checks Response**:
   - ✅ **Success**: If `tool_calls` array is present in response → Function calling supported!
   - ⚠️ **Warning**: If no `tool_calls` → Function calling may not be supported or model chose not to use tools

## Expected Output

### If Function Calling is Supported:

```
✅ FUNCTION CALLING SUPPORTED!
📥 Tool Calls Count: 2

  Tool Call #1:
    ID: call_abc123
    Type: function
    Function Name: update_company_info
    Function Arguments: {"industry":"Kuliner","city":"Jakarta"}

  Tool Call #2:
    ID: call_xyz789
    Type: function
    Function Name: create_product
    Function Arguments: {"name":"Nasi Goreng Spesial","unit_price":25000}
```

### If Function Calling is NOT Supported:

```
⚠️  No tool calls in response
   This could mean:
   1. Function calling not supported by API
   2. Model chose not to use tools (check finish_reason)
   3. API ignored tools parameter
```

## Interpreting Results

### Kolosal.ai (GLM 4.6)
- **If tool_calls present**: ✅ Native function calling works! Use `tools` parameter in requests
- **If no tool_calls**: ⚠️ May need to use structured output parsing fallback

### OpenRouter (x-ai/grok-4-fast)
- **If tool_calls present**: ✅ Function calling works! Use OpenAI-compatible format
- **If no tool_calls**: ⚠️ Check if model supports function calling (grok-4-fast should support it)

## Next Steps

After running the test:

1. **If function calling works**: Update feature brief and proceed with native implementation
2. **If function calling doesn't work**: Use structured output parsing fallback approach
3. **Document results**: Update `specs/active/ai-tools-data-collection/feature-brief.md` with test results

## Troubleshooting

### "API Error 401"
- Check your API key is correct
- Verify API key has proper permissions

### "API Error 400"
- Check request format matches API specification
- Verify `tools` parameter format is correct

### "No tool calls in response"
- Try setting `tool_choice: "required"` to force tool use
- Check if model supports function calling
- Verify tool definitions are correct
