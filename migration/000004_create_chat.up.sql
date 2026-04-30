CREATE TABLE chat_messages (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    sender_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id UUID REFERENCES users(id) ON DELETE CASCADE,
    board_id    UUID REFERENCES boards(id) ON DELETE CASCADE,
    content     TEXT NOT NULL,
    read_at     BIGINT
);

CREATE INDEX idx_chat_dm      ON chat_messages (sender_id, receiver_id, id DESC);
CREATE INDEX idx_chat_game    ON chat_messages (board_id, id ASC);
