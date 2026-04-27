CREATE TABLE chat_messages (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    sender_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id UUID REFERENCES users(id) ON DELETE CASCADE,
    board_id    UUID REFERENCES boards(id) ON DELETE CASCADE,
    content     TEXT NOT NULL,
    read_at     BIGINT,
    created_at  BIGINT NOT NULL
);

CREATE INDEX idx_chat_dm      ON chat_messages (sender_id, receiver_id, created_at DESC);
CREATE INDEX idx_chat_game    ON chat_messages (board_id, created_at ASC);
