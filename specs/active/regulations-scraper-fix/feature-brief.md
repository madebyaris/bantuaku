# Regulations Scraper v2 - AI-Powered UMKM Regulation Discovery

## 🎯 Context (2min)
**Problem**: Current regulations scraper is non-functional - uses wrong URL pattern, never triggered, and database has 0 regulations. UMKM users need relevant Indonesian government regulations for compliance guidance. The RAG system has no data to search.

**Users**: 
- UMKM business owners (need compliance guidance)
- AI Assistant / RAG system (retrieve regulations for insights)
- System administrators (trigger and monitor scraping)

**Success**: Regulations KB populated with UMKM-relevant Indonesian regulations, searchable via RAG, with AI-generated summaries in Bahasa Indonesia.

## 🔍 Quick Research (15min)

### Existing Infrastructure
- **Exa.ai Client** (`backend/services/exa/client.go`) - Already integrated, can search web
- **OpenRouter Embeddings** (`backend/services/embedding/openrouter.go`) - Already configured
- **Chat Provider** - Can generate keywords and summaries via AI
- **Database Tables** - `regulations`, `regulation_chunks`, `regulation_embeddings` exist but empty
- **PDF Extractor** (`backend/services/scraper/regulations/extract.go`) - Has OCR via Kolosal

### New Approach (vs Original)
| Original | Enhanced v2 |
|----------|-------------|
| Crawl peraturan.go.id only | Multi-source: Exa.ai + official site |
| Manual keyword selection | AI generates UMKM keywords (Bahasa) |
| Store PDF URLs | Download PDF → OCR → AI Summary → Delete |
| No fallback for missing PDFs | Exa.ai search for regulation content |
| No embeddings | Generate embeddings for RAG |

## ✅ Requirements

### Core Requirements

1. **AI Keyword Generation**
   - Generate 20-30 UMKM-relevant search keywords in Bahasa Indonesia
   - Categories: pajak, perizinan, ketenagakerjaan, keamanan pangan, HAKI, ekspor/impor
   - Use OpenRouter AI to create contextual keywords

2. **Multi-Source Content Discovery**
   - **Exa.ai Search**: Search for regulation articles using generated keywords
   - **Official Site**: Crawl peraturan.go.id for official regulation PDFs
   - Deduplicate across sources

3. **Content Processing**
   - **With PDF**: Download → Extract text (or OCR if scanned) → AI Summary → Delete PDF
   - **Without PDF**: Use Exa.ai content directly → AI Summary
   - Generate both full text and concise summary

4. **Storage & Embedding**
   - Store regulation metadata + content in `regulations` table
   - Chunk content into `regulation_chunks`
   - Generate embeddings via OpenRouter → store in `embeddings` table
   - Link via `regulation_embeddings`

5. **UMKM Relevance Filtering**
   - AI categorizes regulations by UMKM relevance
   - Categories: `pajak`, `perizinan`, `ketenagakerjaan`, `pangan`, `haki`, `ekspor_impor`, `umum`

## 🏗️ Implementation

### Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                    Regulation Scraper v2                     │
├─────────────────────────────────────────────────────────────┤
│  1. KeywordGenerator                                         │
│     └─ AI → UMKM keywords (Bahasa Indonesia)                 │
│                                                              │
│  2. ContentDiscovery                                         │
│     ├─ ExaSearch → regulation articles                       │
│     └─ OfficialCrawler → peraturan.go.id                     │
│                                                              │
│  3. ContentProcessor                                         │
│     ├─ PDFProcessor: Download → OCR → Summary → Delete       │
│     └─ WebProcessor: Exa content → Summary                   │
│                                                              │
│  4. StorageService                                           │
│     └─ Store → Chunk → Embed → Link                          │
└─────────────────────────────────────────────────────────────┘
```

### Files to Modify/Create
- `backend/services/scraper/regulations/keywords.go` (NEW) - AI keyword generator
- `backend/services/scraper/regulations/discovery.go` (NEW) - Multi-source discovery
- `backend/services/scraper/regulations/processor.go` (NEW) - Content processor
- `backend/services/scraper/regulations/scheduler.go` - Update job orchestration
- `backend/services/scraper/regulations/store.go` - Add embedding integration

### Dependencies
- Exa.ai client (existing)
- OpenRouter chat provider (existing)
- OpenRouter embeddings (existing)
- Kolosal OCR (existing)

## 📋 Next Actions

1. **Implement KeywordGenerator** (30min)
   - Create `keywords.go` with AI-powered keyword generation
   - Generate UMKM-specific keywords in Bahasa Indonesia

2. **Implement Exa Search Integration** (1h)
   - Create `discovery.go` with Exa.ai search
   - Search for regulations using generated keywords

3. **Implement Content Processor** (1h)
   - Create `processor.go` for PDF and web content
   - PDF: Download → OCR → AI Summary → Delete
   - Web: Extract content → AI Summary

4. **Update Storage with Embeddings** (1h)
   - Modify `store.go` to generate and store embeddings
   - Chunk content properly for RAG

5. **Update Scheduler** (30min)
   - Orchestrate new pipeline
   - Add job status tracking

6. **Test End-to-End** (30min)
   - Trigger scraper
   - Verify regulations in DB
   - Test RAG search

---
**Total Planning Time**: ~30min | **Owner**: Development Team | 2025-12-06

## Changelog

### 2025-12-06 - Major Architecture Overhaul
**Change:** Complete redesign from simple crawler to AI-powered multi-source discovery
**Reason:** Original approach non-functional, need AI assistance for keyword generation and content summarization
**Impact:** New files created, existing scheduler rewritten, embedding integration added

### Key Additions:
- AI keyword generation (Bahasa Indonesia)
- Exa.ai as alternative content source
- PDF → OCR → AI Summary → Delete pipeline
- Embedding generation for RAG
- UMKM relevance categorization

---
**See**: [todo-list.md](./todo-list.md) for detailed implementation tasks
