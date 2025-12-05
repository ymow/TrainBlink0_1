-- Rollback chat_messages table creation
-- Phase 1: Basic Messaging System
-- Created: 2025-12-05

DROP TRIGGER IF EXISTS messages_updated_at_trigger ON chat_messages;
DROP FUNCTION IF EXISTS update_messages_updated_at();
DROP TABLE IF EXISTS chat_messages;
