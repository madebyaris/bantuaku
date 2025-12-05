package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bantuaku/backend/config"
	"github.com/bantuaku/backend/services/storage"
)

func main() {
	// Load config
	cfg := config.Load()

	// Connect to database
	db, err := storage.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()
	demoCompanyID := "demo-store-001"

	fmt.Printf("Resetting chat for demo account (company_id: %s)...\n", demoCompanyID)

	// Delete all messages for demo company conversations
	result, err := db.Pool().Exec(ctx, `
		DELETE FROM messages
		WHERE conversation_id IN (
			SELECT id FROM conversations
			WHERE company_id = $1
		)
	`, demoCompanyID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to delete messages: %v\n", err)
		os.Exit(1)
	}
	messagesDeleted := result.RowsAffected()
	fmt.Printf("Deleted %d messages\n", messagesDeleted)

	// Delete all conversations for demo company
	result, err = db.Pool().Exec(ctx, `
		DELETE FROM conversations
		WHERE company_id = $1
	`, demoCompanyID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to delete conversations: %v\n", err)
		os.Exit(1)
	}
	conversationsDeleted := result.RowsAffected()
	fmt.Printf("Deleted %d conversations\n", conversationsDeleted)

	fmt.Printf("\n✅ Chat reset complete for demo account!\n")
	fmt.Printf("   - Messages deleted: %d\n", messagesDeleted)
	fmt.Printf("   - Conversations deleted: %d\n", conversationsDeleted)
}
