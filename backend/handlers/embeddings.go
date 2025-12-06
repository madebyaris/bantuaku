package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/embedding"
)

// IndexChunks triggers batch indexing of regulation chunks
func (h *Handler) IndexChunks(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))

	// Get limit query parameter (default: 100, max: 10000)
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 10000 {
			limit = parsed
		}
	}

	// Create embedder
	embedder, err := embedding.NewEmbedder(h.config)
	if err != nil {
		log.Error("Failed to create embedder", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to initialize embedding service"), r)
		return
	}

	// Create indexer
	indexer := embedding.NewIndexer(h.db.Pool(), embedder)

	// Run indexing in goroutine with detached context (request context will be cancelled)
	go func() {
		// Create a new logger for the goroutine without request-scoped values
		bgLog := logger.With("job", "index_chunks", "limit", limit)
		// Use background context with timeout since request context will be cancelled
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		count, err := indexer.IndexChunks(ctx, limit)
		if err != nil {
			bgLog.Error("Indexing failed", "error", err)
		} else {
			bgLog.Info("Indexing completed", "indexed", count)
		}
	}()

	log.Info("Chunk indexing triggered", "limit", limit)

	h.respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "Indexing job started",
		"limit":   limit,
		"status":  "running",
	})
}

// SearchRegulations performs semantic search on regulations
func (h *Handler) SearchRegulations(w http.ResponseWriter, r *http.Request) {
	log := logger.With("request_id", r.Context().Value("request_id"))

	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		h.respondError(w, errors.NewValidationError("Query parameter 'q' is required", ""), r)
		return
	}

	// Get k parameter (default: 5)
	k := 5
	if kStr := r.URL.Query().Get("k"); kStr != "" {
		if parsed, err := strconv.Atoi(kStr); err == nil && parsed > 0 && parsed <= 50 {
			k = parsed
		}
	}

	// Parse filters
	var filters embedding.Filters
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filters.Year = &year
		}
	}
	if category := r.URL.Query().Get("category"); category != "" {
		filters.Category = &category
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = &status
	}

	// Create embedder
	embedder, err := embedding.NewEmbedder(h.config)
	if err != nil {
		log.Error("Failed to create embedder", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Failed to initialize embedding service"), r)
		return
	}

	// Create retrieval service
	retrieval := embedding.NewRetrievalService(h.db.Pool(), embedder)

	// Perform search
	chunks, err := retrieval.Search(r.Context(), query, k, filters)
	if err != nil {
		log.Error("Search failed", "error", err)
		h.respondError(w, errors.NewInternalError(err, "Search failed"), r)
		return
	}

	log.Info("Search completed", "query", query, "results", len(chunks))

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"query":   query,
		"results": chunks,
		"count":   len(chunks),
	})
}
