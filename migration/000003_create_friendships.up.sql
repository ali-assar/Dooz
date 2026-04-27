CREATE TABLE friendships (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    requester_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    addressee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       SMALLINT NOT NULL DEFAULT 1,
        -- 1=pending 2=accepted 3=rejected 4=blocked
    created_at   BIGINT NOT NULL,
    updated_at   BIGINT NOT NULL,

    UNIQUE (requester_id, addressee_id)
);

CREATE INDEX idx_friendships_addressee ON friendships (addressee_id, status);
CREATE INDEX idx_friendships_requester ON friendships (requester_id, status);
