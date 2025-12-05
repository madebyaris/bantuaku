package handlers

import (
	"context"
	"fmt"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/embedding"
)

// RAGService handles RAG integration for chat
type RAGService struct {
	retrieval *embedding.RetrievalService
	log       logger.Logger
}

// NewRAGService creates a new RAG service
func NewRAGService(retrieval *embedding.RetrievalService) *RAGService {
	log := logger.Default()
	return &RAGService{
		retrieval: retrieval,
		log:       *log,
	}
}

// BuildRAGContext builds context from retrieved regulation chunks
func (r *RAGService) BuildRAGContext(ctx context.Context, query string, k int) (string, []embedding.RetrievedChunk, error) {
	// Perform retrieval
	chunks, err := r.retrieval.Search(ctx, query, k, embedding.Filters{})
	if err != nil {
		r.log.Warn("RAG retrieval failed", "error", err)
		return "", nil, fmt.Errorf("failed to retrieve regulations: %w", err)
	}

	if len(chunks) == 0 {
		return "", nil, nil
	}

	// Format chunks for context
	contextText := embedding.FormatChunksForContext(chunks)

	return contextText, chunks, nil
}

// BuildRAGPrompt builds a prompt with RAG context
func (r *RAGService) BuildRAGPrompt(userQuery string, ragContext string) (string, string) {
	systemPrompt := `Kamu adalah Asisten Bantuaku, AI assistant untuk membantu UMKM Indonesia dengan informasi tentang peraturan pemerintah Indonesia.

Instruksi:
1. Jawab dalam Bahasa Indonesia yang natural dan ramah
2. Gunakan informasi dari konteks peraturan yang diberikan untuk menjawab pertanyaan
3. Jika informasi tidak tersedia dalam konteks, katakan dengan jujur bahwa kamu tidak memiliki informasi tersebut
4. Selalu sertakan referensi ke peraturan yang digunakan (nomor peraturan, tahun)
5. Berikan jawaban yang praktis dan dapat ditindaklanjuti`

	userPrompt := userQuery

	if ragContext != "" {
		userPrompt = fmt.Sprintf(`Konteks Peraturan Pemerintah Indonesia:

%s

Pertanyaan Pengguna: %s

Jawab pertanyaan di atas menggunakan informasi dari konteks peraturan yang diberikan. Jika informasi tidak tersedia dalam konteks, katakan dengan jujur.`, ragContext, userQuery)
	}

	return systemPrompt, userPrompt
}

// ExtractCitations extracts citation information from retrieved chunks
func ExtractCitations(chunks []embedding.RetrievedChunk) []Citation {
	citations := make([]Citation, 0, len(chunks))
	seen := make(map[string]bool)

	for _, chunk := range chunks {
		// Use regulation ID as key to avoid duplicates
		key := chunk.RegulationID
		if seen[key] {
			continue
		}
		seen[key] = true

		citation := Citation{
			Title:           chunk.RegulationTitle,
			RegulationNumber: chunk.RegulationNumber,
			Year:            chunk.Year,
			Category:       chunk.Category,
			SourceURL:      chunk.SourceURL,
			PDFURL:         chunk.PDFURL,
			SectionNumber:  chunk.SectionNumber,
			SectionTitle:   chunk.SectionTitle,
		}
		citations = append(citations, citation)
	}

	return citations
}

// Citation represents a citation to a regulation
type Citation struct {
	Title            string  `json:"title"`
	RegulationNumber string  `json:"regulation_number"`
	Year             int     `json:"year"`
	Category         string  `json:"category"`
	SourceURL        string  `json:"source_url"`
	PDFURL           string  `json:"pdf_url"`
	SectionNumber    *string `json:"section_number,omitempty"`
	SectionTitle     *string `json:"section_title,omitempty"`
}

