package admin

import (
	"net/http"
	"strconv"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/services/transactions"
)

// GetSubscriptionTransactions retrieves transaction history for a subscription
func (h *AdminHandler) GetSubscriptionTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract subscription ID from path
	path := r.URL.Path
	subscriptionID := path[len("/api/v1/admin/subscriptions/"):]
	if idx := len(subscriptionID) - len("/transactions"); idx > 0 && subscriptionID[idx:] == "/transactions" {
		subscriptionID = subscriptionID[:idx]
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get transaction history
	transactionService := transactions.NewService(h.db)
	txns, total, err := transactionService.GetTransactionHistory(ctx, subscriptionID, page, limit)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "get transaction history")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": txns,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
