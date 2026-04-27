CREATE TABLE achievements (
    id                UUID PRIMARY KEY DEFAULT uuidv7(),
    name              TEXT NOT NULL UNIQUE,
    description       TEXT NOT NULL,
    icon              TEXT NOT NULL DEFAULT '',
    requirement_type  SMALLINT NOT NULL,
        -- 1=wins 2=draws 3=friends 4=games_played 5=win_streak
    requirement_value INTEGER NOT NULL,
    coin_reward       INTEGER NOT NULL DEFAULT 0,
    gem_reward        INTEGER NOT NULL DEFAULT 0,
    created_at        BIGINT NOT NULL,
    updated_at        BIGINT NOT NULL
);

CREATE TABLE user_achievements (
    id             UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    earned_at      BIGINT NOT NULL,

    UNIQUE (user_id, achievement_id)
);

-- Seed initial achievements
INSERT INTO achievements (id, name, description, icon, requirement_type, requirement_value, coin_reward, gem_reward, created_at, updated_at) VALUES
    (uuidv7(), 'First Win',        'Win your first game',       '🏆', 1, 1,   50,  0,  extract(epoch from now())::bigint, extract(epoch from now())::bigint),
    (uuidv7(), 'Hat Trick',        'Win 3 games',               '🎩', 1, 3,   100, 0,  extract(epoch from now())::bigint, extract(epoch from now())::bigint),
    (uuidv7(), 'Victor',           'Win 10 games',              '🥇', 1, 10,  200, 1,  extract(epoch from now())::bigint, extract(epoch from now())::bigint),
    (uuidv7(), 'Champion',         'Win 50 games',              '👑', 1, 50,  500, 5,  extract(epoch from now())::bigint, extract(epoch from now())::bigint),
    (uuidv7(), 'Peacemaker',       'Draw 5 games',              '🤝', 2, 5,   75,  0,  extract(epoch from now())::bigint, extract(epoch from now())::bigint),
    (uuidv7(), 'Social Butterfly', 'Add 5 friends',             '🦋', 3, 5,   100, 0,  extract(epoch from now())::bigint, extract(epoch from now())::bigint),
    (uuidv7(), 'Veteran',          'Play 25 games',             '🎖️', 4, 25,  150, 0,  extract(epoch from now())::bigint, extract(epoch from now())::bigint),
    (uuidv7(), 'Win Streak',       'Win 3 games in a row',      '🔥', 5, 3,   200, 2,  extract(epoch from now())::bigint, extract(epoch from now())::bigint);
