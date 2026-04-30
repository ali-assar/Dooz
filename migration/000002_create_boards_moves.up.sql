CREATE TABLE boards (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    player_x_id     UUID NOT NULL REFERENCES users(id),
    player_o_id     UUID REFERENCES users(id),
    winner_id       UUID REFERENCES users(id),
    status          SMALLINT NOT NULL DEFAULT 1,
        -- 1=waiting 2=in_progress 3=completed 4=abandoned
    is_bot_game     BOOLEAN NOT NULL DEFAULT false,
    board_state     CHAR(9) NOT NULL DEFAULT '---------',
    current_turn    UUID REFERENCES users(id),
    started_at      BIGINT,
    ended_at        BIGINT,
    updated_at      BIGINT NOT NULL
);

CREATE INDEX idx_boards_player_x ON boards (player_x_id);
CREATE INDEX idx_boards_player_o ON boards (player_o_id);
CREATE INDEX idx_boards_status   ON boards (status);

CREATE TABLE moves (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    board_id    UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id),
    position    SMALLINT NOT NULL,  -- 0-8
    symbol      CHAR(1) NOT NULL    -- 'X' or 'O'
);

CREATE INDEX idx_moves_board ON moves (board_id, id ASC);
