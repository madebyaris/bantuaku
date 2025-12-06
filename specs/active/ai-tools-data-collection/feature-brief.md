# AI Tools & Data Collection Feature Brief

## 🎯 Context (2min)
**Problem**: AI assistant cannot proactively collect missing company/product data during chat conversations. Users must manually input data through forms, creating friction and incomplete profiles.

**Users**: 
- UMKM owners interacting with AI assistant via chat
- Users who prefer conversational data entry over forms

**Success**: 
- AI detects missing data and asks users naturally via chat
- AI can invoke tools to store collected data automatically
- Users complete company/product profiles entirely through conversation
- File uploads for forecasts handled seamlessly in chat context

## 🔍 Quick Research (15min)

### Existing Patterns

#### Chat System ✅
- **Location**: `backend/handlers/chat.go`, `backend/services/chat/`
- **Pattern**: Kolosal.ai/OpenRouter integration via `ChatProvider` interface
- **Usage**: RAG-enhanced prompts, conversation persistence
- **Reuse**: Extend `SendMessage` handler for tool execution

#### Product Creation API ✅
- **Location**: `backend/handlers/products.go` → `CreateProduct`
- **Endpoint**: `POST /api/v1/products`
- **Pattern**: Validates `company_id` from context, creates product record
- **Reuse**: Can be called internally by tool executor

#### Company Model ✅
- **Location**: `backend/models/company.go`
- **Fields**: `Industry`, `BusinessModel`, `LocationRegion`, `City`, `SocialMediaHandles` (JSONB)
- **Pattern**: Company created during registration, can be updated
- **Reuse**: Need update endpoint or create internal update function

#### File Upload System ✅
- **Location**: `backend/handlers/files.go` → `UploadFile`
- **Endpoint**: `POST /api/v1/files/upload`
- **Pattern**: Multipart form, supports CSV/XLSX/PDF, OCR via Kolosal.ai
- **Reuse**: Can be invoked by tool, but needs chat context integration

#### System Prompt Structure ✅
- **Location**: `backend/handlers/chat.go` (line 231), `backend/handlers/rag.go` (line 47)
- **Pattern**: Basic prompt + RAG context injection
- **Reuse**: Extend with tool usage instructions

### Model Research Findings

**Current Model**: GLM-4.6 (Zhipu AI) via Kolosal.ai
- **Model Name**: "GLM 4.6" (as used in codebase: `backend/handlers/chat.go:248`)
- **Provider**: Kolosal.ai API (`https://api.kolosal.ai/v1/chat/completions`)

**GLM-4.6 Function Calling Capabilities** ✅:
- ✅ **Native Function Calling Support**: GLM-4.6 natively supports function calling
- ✅ **OpenAI-Compatible Format**: Uses `tools` parameter (array of function definitions)
- ✅ **Tool Choice Control**: `tool_choice` parameter (`none`, `auto`, `required`)
- ✅ **Streaming Support**: Supports streaming tool calls for real-time responses
- ✅ **Extended Context**: 200K token context window (up from 128K)
- ✅ **Enhanced Reasoning**: Improved tool use during inference for agentic applications

**API Format** (Expected):
```json
{
  "model": "GLM 4.6",
  "messages": [...],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "function_name",
        "description": "...",
        "parameters": {...}
      }
    }
  ],
  "tool_choice": "auto"
}
```

**Response Format** (Expected):
```json
{
  "choices": [{
    "message": {
      "tool_calls": [{
        "id": "call_...",
        "type": "function",
        "function": {
          "name": "function_name",
          "arguments": "{...}"
        }
      }]
    },
    "finish_reason": "tool_calls"
  }]
}
```

**Verification Status**: 
- ⚠️ **Kolosal.ai API Documentation**: Not publicly available for function calling
- ⚠️ **Testing Required**: Need to test if Kolosal.ai exposes GLM-4.6's function calling via API
- ✅ **Model Support Confirmed**: GLM-4.6 definitely supports function calling

### Tech Decisions

**Function Calling Approach**: 
- **Primary Strategy**: Native function calling via `tools` parameter (GLM-4.6 capability)
- **Decision**: **Start with Option A (Native Function Calling)** - Test Kolosal.ai API with `tools` parameter
- **Fallback**: If Kolosal.ai doesn't expose function calling, use Option B (structured output parsing)
- **Why**: GLM-4.6 has native support, better performance, cleaner implementation, streaming support
- **Testing Required**: Verify Kolosal.ai API accepts `tools` parameter (test API call or contact Kolosal support)

**Tool Execution Pattern**:
- **Approach**: Internal service layer that AI tools can call
- **Why**: Reuses existing handlers, maintains security (company_id from context), consistent error handling
- **Avoid**: Direct database access from tool executor

**Prompt Engineering**:
- **Approach**: Multi-step system prompt that guides AI through data collection workflow
- **Why**: Ensures AI asks for missing data systematically, uses tools appropriately
- **Avoid**: Overly complex prompts that confuse the model

## ✅ Requirements (10min)

### Core User Stories

1. **AI Detects Missing Company Data**
   - **Story**: AI checks company profile, identifies missing fields (type, location, social media)
   - **Acceptance**: AI asks user "Apa jenis bisnis Anda?" if `industry` is empty
   - **Tool**: `check_company_profile` → returns missing fields list

2. **AI Collects Company Type/Industry**
   - **Story**: User responds with business type, AI stores it
   - **Acceptance**: AI calls `update_company_info` with `industry` field
   - **Tool**: `update_company_info(industry: string, business_model?: string)`

3. **AI Collects Location Data**
   - **Story**: AI asks "Di kota mana bisnis Anda berada?", stores location
   - **Acceptance**: AI calls `update_company_info` with `city` and `location_region`
   - **Tool**: `update_company_info(city: string, location_region?: string)`

4. **AI Collects Social Media**
   - **Story**: AI asks "Apa akun Instagram/TikTok/Tokopedia Anda?", stores handles
   - **Acceptance**: AI calls `update_company_social_media` with platform → handle mapping
   - **Tool**: `update_company_social_media(platform: string, handle: string)`

5. **AI Requests Forecast File Upload**
   - **Story**: AI asks "Bisakah Anda upload file forecast penjualan (PDF/XLSX/CSV)?"
   - **Acceptance**: User provides file, AI processes and stores forecast data
   - **Tool**: `request_file_upload(type: "forecast", description: string)` → returns upload URL/token

6. **AI Processes Uploaded Forecast File**
   - **Story**: After file upload, AI extracts data and creates forecast records
   - **Acceptance**: AI calls `process_forecast_file(file_id: string)` → extracts products/sales
   - **Tool**: `process_forecast_file(file_id: string)` → returns extracted data summary

7. **AI Collects Product/Service Information**
   - **Story**: AI asks "Apa produk atau layanan yang Anda jual?", creates product records
   - **Acceptance**: AI calls `create_product` for each product mentioned
   - **Tool**: `create_product(name: string, category?: string, unit_price?: number, sku?: string)`

8. **AI Validates Before Storing**
   - **Story**: AI confirms data before calling tools ("Apakah nama produknya 'Nasi Goreng Spesial'?")
   - **Acceptance**: AI asks for confirmation on critical fields, only stores after user confirms
   - **Tool**: No tool needed, handled in prompt logic

### Tool Definitions

**Tool 1: `check_company_profile`**
```json
{
  "name": "check_company_profile",
  "description": "Check current company profile and identify missing required fields",
  "parameters": {}
}
```

**Tool 2: `update_company_info`**
```json
{
  "name": "update_company_info",
  "description": "Update company information fields (industry, location, business model)",
  "parameters": {
    "type": "object",
    "properties": {
      "industry": {"type": "string", "description": "Business industry/type"},
      "business_model": {"type": "string", "description": "Business model (e.g., 'retail', 'service', 'manufacturing')"},
      "city": {"type": "string", "description": "City where business operates"},
      "location_region": {"type": "string", "description": "Region/province"},
      "description": {"type": "string", "description": "Company description"}
    }
  }
}
```

**Tool 3: `update_company_social_media`**
```json
{
  "name": "update_company_social_media",
  "description": "Add or update social media handles for the company",
  "parameters": {
    "type": "object",
    "properties": {
      "platform": {"type": "string", "enum": ["instagram", "tiktok", "facebook", "twitter", "tokopedia", "shopee", "lazada", "bukalapak"]},
      "handle": {"type": "string", "description": "Social media handle/username (without @)"}
    },
    "required": ["platform", "handle"]
  }
}
```

**Tool 4: `request_file_upload`**
```json
{
  "name": "request_file_upload",
  "description": "Request user to upload a file (forecast, sales data, etc.)",
  "parameters": {
    "type": "object",
    "properties": {
      "file_type": {"type": "string", "enum": ["forecast", "sales", "products"], "description": "Type of file expected"},
      "description": {"type": "string", "description": "What to ask user for (e.g., 'forecast penjualan bulan depan')"},
      "accepted_formats": {"type": "array", "items": {"type": "string"}, "default": ["pdf", "xlsx", "csv"]}
    },
    "required": ["file_type", "description"]
  }
}
```

**Tool 5: `process_forecast_file`**
```json
{
  "name": "process_forecast_file",
  "description": "Process an uploaded forecast file and extract product/sales data",
  "parameters": {
    "type": "object",
    "properties": {
      "file_id": {"type": "string", "description": "File upload ID from request_file_upload response"}
    },
    "required": ["file_id"]
  }
}
```

**Tool 6: `create_product`**
```json
{
  "name": "create_product",
  "description": "Create a new product or service for the company",
  "parameters": {
    "type": "object",
    "properties": {
      "name": {"type": "string", "description": "Product/service name"},
      "category": {"type": "string", "description": "Product category"},
      "unit_price": {"type": "number", "description": "Price per unit"},
      "cost": {"type": "number", "description": "Cost per unit"},
      "sku": {"type": "string", "description": "SKU code"}
    },
    "required": ["name"]
  }
}
```

**Tool 7: `list_products`**
```json
{
  "name": "list_products",
  "description": "List all products/services for the company",
  "parameters": {
    "type": "object",
    "properties": {
      "category": {"type": "string", "description": "Filter by category (optional)"}
    }
  }
}
```

## 🏗️ Implementation (5min)

### Components

**1. Tool Executor Service** (`backend/services/tools/executor.go`)
- Receives tool name + parameters from AI response
- Maps to internal handler calls
- Returns structured results to AI
- Handles errors gracefully

**2. Enhanced System Prompt Builder** (`backend/handlers/chat.go`)
- Detects missing company/product data
- Injects tool definitions into system prompt
- Guides AI on when to use tools
- Includes data collection workflow instructions

**3. Function Calling Handler** (`backend/handlers/chat.go`)
- **Primary**: Parses native function calling response (`tool_calls` array)
- **Fallback**: Parses structured output JSON if native unavailable
- Extracts tool name and parameters
- Validates tool calls before execution

**4. Company Update Handler** (`backend/handlers/companies.go` - NEW)
- `PATCH /api/v1/companies/me` endpoint
- Updates company fields (industry, location, social media)
- Used by tool executor

**5. Tool Response Formatter** (`backend/services/tools/formatter.go`)
- Formats tool execution results for AI
- Includes success/error messages
- Provides context for next AI turn

### APIs

**New Endpoints**:
- `PATCH /api/v1/companies/me` - Update company info (internal + external)
- `POST /api/v1/chat/tools/execute` - Execute tool from AI (internal)

**Enhanced Endpoints**:
- `POST /api/v1/chat/messages` - Now handles function calling/tool calls in response
  - Sends `tools` parameter to Kolosal.ai API (if supported)
  - Parses `tool_calls` from response
  - Executes tools and sends results back to AI
- `POST /api/v1/files/upload` - Enhanced to return file_id for tool processing

**Reused Endpoints**:
- `POST /api/v1/products` - Create product (via tool executor)
- `GET /api/v1/products` - List products (via tool executor)

### Data Changes

**No Schema Changes Required** ✅
- All fields exist in `companies` table
- `products` table supports all needed fields
- `file_uploads` table exists for forecast files

**New Internal Models**:
- `ToolCall` - Represents AI tool invocation
- `ToolResult` - Represents tool execution result

## 📋 Next Actions (2min)

- [x] Research model capabilities - **COMPLETED**: GLM-4.6 supports native function calling ✅
- [ ] **CRITICAL TEST**: Verify Kolosal.ai API accepts `tools` parameter (test API call with tools) (30min)
  - Create test request with `tools` array
  - Check if response includes `tool_calls` in message
  - Document API behavior (supported/not supported)
- [ ] Design function calling format (OpenAI-compatible `tools` array) (15min)
- [ ] **FALLBACK**: Design structured output format if native function calling unavailable (10min)
- [ ] Create tool executor service with handler mapping (30min)
- [ ] Implement company update endpoint `PATCH /api/v1/companies/me` (20min)
- [ ] Enhance system prompt with tool definitions and data collection workflow (30min)
- [ ] Add function calling support in `SendMessage` handler:
  - Add `tools` parameter to `ChatCompletionRequest` (Kolosal client) (15min)
  - Parse `tool_calls` from response (native format) (20min)
  - Fallback: Parse structured output JSON if native unavailable (15min)
- [ ] Create tool response formatter for AI context (15min)
- [ ] Test AI tool invocation flow end-to-end (30min)

**Start Coding In**: ~3.5 hours (after API testing and design)

**Critical First Step**: Test Kolosal.ai API with `tools` parameter to confirm native function calling support before implementation.

---
**Total Planning Time**: ~30min | **Owner**: TBD | Dec 5, 2025

<!-- Living Document - Update as you code -->

## 🔄 Implementation Tracking

**CRITICAL**: Follow the todo-list systematically. Mark items as complete, document blockers, update progress.

### Progress
- [ ] Research phase complete
- [ ] Tool executor service implemented
- [ ] System prompt enhanced
- [ ] Tool execution integrated into chat flow
- [ ] End-to-end testing complete

### Blockers
- [ ] Document any blockers here

**See**: [.sdd/IMPLEMENTATION_GUIDE.md](mdc:.sdd/IMPLEMENTATION_GUIDE.md) for detailed execution rules.

## 📝 Notes

### System Prompt Enhancement Strategy

The enhanced system prompt should:

1. **Check Company Profile First**: "Sebelum menjawab, periksa profil perusahaan menggunakan `check_company_profile`. Jika ada field yang kosong, tanyakan kepada user."

2. **Data Collection Workflow**: 
   - Ask for company type/industry
   - Ask for location (city, region)
   - Ask for social media handles (one at a time)
   - Ask for products/services
   - Request forecast file if needed

3. **Tool Usage Instructions**:
   - "Gunakan tool `update_company_info` untuk menyimpan informasi perusahaan"
   - "Gunakan tool `create_product` untuk menambahkan produk/layanan"
   - "Gunakan tool `request_file_upload` jika user perlu upload file forecast"

4. **Confirmation Pattern**:
   - "Sebelum menyimpan data, konfirmasi dengan user: 'Apakah [field] adalah [value]?'"
   - "Hanya panggil tool setelah user mengkonfirmasi"

5. **Natural Language**:
   - "Tanyakan dengan ramah dan natural, seperti sedang berbicara dengan teman"
   - "Jika user tidak yakin, berikan contoh atau pilihan"

### Function Calling Format (Native - Preferred)

**If Kolosal.ai supports native function calling** (GLM-4.6 capability):

Request format (OpenAI-compatible):
```json
{
  "model": "GLM 4.6",
  "messages": [...],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "update_company_info",
        "description": "Update company information fields",
        "parameters": {
          "type": "object",
          "properties": {
            "industry": {"type": "string"},
            "city": {"type": "string"}
          }
        }
      }
    }
  ],
  "tool_choice": "auto"  // or "required" to force tool use
}
```

Response format:
```json
{
  "choices": [{
    "message": {
      "role": "assistant",
      "content": null,
      "tool_calls": [{
        "id": "call_abc123",
        "type": "function",
        "function": {
          "name": "update_company_info",
          "arguments": "{\"industry\": \"Kuliner\", \"city\": \"Jakarta\"}"
        }
      }]
    },
    "finish_reason": "tool_calls"
  }]
}
```

### Structured Output Format (Fallback)

**If native function calling unavailable**, AI should return tool calls in this format:

```json
{
  "response": "Baik, saya akan menyimpan informasi perusahaan Anda.",
  "tool_calls": [
    {
      "tool": "update_company_info",
      "parameters": {
        "industry": "Kuliner",
        "city": "Jakarta",
        "location_region": "DKI Jakarta"
      }
    }
  ]
}
```

If no tool call needed:
```json
{
  "response": "Terima kasih atas informasinya!"
}
```

### Error Handling

- If tool execution fails, AI should inform user and ask for clarification
- If required parameter missing, AI should ask user for it
- If file upload fails, AI should guide user to retry or provide alternative
