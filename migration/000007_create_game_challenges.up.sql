CREATE TABLE game_challenges (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    requester_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    addressee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    board_id     UUID REFERENCES boards(id) ON DELETE SET NULL,
    status       SMALLINT NOT NULL DEFAULT 1, -- 1=pending 2=accepted 3=rejected 4=canceled 5=expired
    expires_at   BIGINT NOT NULL,
    CHECK (requester_id <> addressee_id)
);

CREATE INDEX idx_game_challenges_addressee_status ON game_challenges (addressee_id, status, expires_at);
CREATE INDEX idx_game_challenges_requester_status ON game_challenges (requester_id, status, expires_at);

CREATE UNIQUE INDEX uq_active_challenge_pair
ON game_challenges (LEAST(requester_id, addressee_id), GREATEST(requester_id, addressee_id))
WHERE status = 1;
