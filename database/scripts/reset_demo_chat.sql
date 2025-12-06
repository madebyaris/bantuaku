-- Reset Chat for Demo Account
-- This script deletes all conversations and messages for the demo account
-- Run with: psql -U your_user -d bantuaku -f database/scripts/reset_demo_chat.sql

-- Delete all messages for demo company conversations
DELETE FROM messages
WHERE conversation_id IN (
    SELECT id FROM conversations
    WHERE company_id = 'demo-store-001'
);

-- Delete all conversations for demo company
DELETE FROM conversations
WHERE company_id = 'demo-store-001';

-- Verify deletion
SELECT 
    (SELECT COUNT(*) FROM conversations WHERE company_id = 'demo-store-001') as remaining_conversations,
    (SELECT COUNT(*) FROM messages WHERE conversation_id IN (
        SELECT id FROM conversations WHERE company_id = 'demo-store-001'
    )) as remaining_messages;

-- Output success message
DO $$
BEGIN
    RAISE NOTICE 'Chat reset complete for demo account (demo-store-001)';
END $$;
