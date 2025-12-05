-- Create chat_messages table for PostgreSQL
-- Phase 1: Basic Messaging System
-- Created: 2025-12-05

CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Content
    text TEXT NOT NULL,

    -- Participants
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Timing
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Delivery tracking
    delivery_status VARCHAR(20) DEFAULT 'PENDING'
        CHECK (delivery_status IN ('PENDING', 'SENDING', 'SENT', 'DELIVERED', 'READ', 'FAILED')),
    delivered_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    is_read BOOLEAN DEFAULT false,

    -- Encryption
    is_encrypted BOOLEAN DEFAULT false,

    -- Ephemeral messages (Feature 6 - future)
    is_ephemeral BOOLEAN DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_expired BOOLEAN DEFAULT false,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- CRITICAL INDEXES for query performance
-- Index for sender -> receiver queries
CREATE INDEX idx_messages_sender_receiver ON chat_messages(sender_id, receiver_id, timestamp DESC);

-- Index for receiver -> sender queries
CREATE INDEX idx_messages_receiver_sender ON chat_messages(receiver_id, sender_id, timestamp DESC);

-- Index for bidirectional conversation queries
-- Uses LEAST/GREATEST to normalize participant order
CREATE INDEX idx_messages_conversation ON chat_messages(
    LEAST(sender_id, receiver_id),
    GREATEST(sender_id, receiver_id),
    timestamp DESC
);

-- Index for delivery status filtering
CREATE INDEX idx_messages_status ON chat_messages(delivery_status);

-- Partial index for unread messages (more efficient than full index)
CREATE INDEX idx_messages_unread ON chat_messages(receiver_id, is_read, timestamp DESC)
    WHERE is_read = false;

-- Updated_at trigger to automatically update timestamp on row modification
CREATE OR REPLACE FUNCTION update_messages_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER messages_updated_at_trigger
    BEFORE UPDATE ON chat_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_messages_updated_at();
