CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    email           TEXT NOT NULL UNIQUE,
    fullname        TEXT NOT NULL,
    password        TEXT,
    phone           TEXT UNIQUE,
    user_code       INTEGER NOT NULL UNIQUE CHECK (user_code BETWEEN 100000 AND 999999),
    role            SMALLINT NOT NULL DEFAULT 1,  -- 1=user 2=admin 3=super_admin
    is_email_verified BOOLEAN NOT NULL DEFAULT false,
    is_phone_verified BOOLEAN NOT NULL DEFAULT false,
    current_theme   SMALLINT NOT NULL DEFAULT 1,
    current_xo_shape SMALLINT NOT NULL DEFAULT 1,
    current_avatar  SMALLINT NOT NULL DEFAULT 1,
    coins           INTEGER NOT NULL DEFAULT 0,
    gems            INTEGER NOT NULL DEFAULT 0,
    x_count         INTEGER NOT NULL DEFAULT 0,
    o_count         INTEGER NOT NULL DEFAULT 0,
    wins            INTEGER NOT NULL DEFAULT 0,
    losses          INTEGER NOT NULL DEFAULT 0,
    draws           INTEGER NOT NULL DEFAULT 0,
    is_online       BOOLEAN NOT NULL DEFAULT false,
    last_seen_at    BIGINT,
    updated_at      BIGINT NOT NULL,
    deleted_at      BIGINT,
    blocked_at      BIGINT
);

CREATE TABLE user_sessions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id             UUID NOT NULL,
    session_token       TEXT,
    jwt_id              TEXT UNIQUE,
    refresh_token_hash  TEXT NOT NULL,
    expires_at          BIGINT NOT NULL,
    ip_address          INET,
    user_agent          TEXT,
    device_type         SMALLINT NOT NULL,  -- 1=web 2=mobile 3=telegram
    last_activity_at    BIGINT NOT NULL,

    UNIQUE (user_id, device_type),
    CONSTRAINT fk_user_sessions_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE otp_outbox (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    recipient     TEXT NOT NULL,
    channel       SMALLINT NOT NULL DEFAULT 1,   -- 1=SMS 2=Email
    otp_code      TEXT NOT NULL,
    expires_at    BIGINT NOT NULL,
    purpose       SMALLINT NOT NULL,             -- 1=registration 2=forgot_password
    status        SMALLINT NOT NULL DEFAULT 0,   -- 0=pending 1=sent 2=failed
    retry_count   SMALLINT NOT NULL DEFAULT 0,
    next_retry_at BIGINT,
    processed_at  BIGINT
);
