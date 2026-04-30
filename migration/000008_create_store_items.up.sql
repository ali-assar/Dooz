CREATE TABLE store_items (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    item_type       SMALLINT NOT NULL, -- 1=theme 2=xo_shape 3=avatar
    item_value      SMALLINT NOT NULL,
    item_key        TEXT NOT NULL UNIQUE,
    asset_url       TEXT NOT NULL DEFAULT '',
    price_currency  SMALLINT NOT NULL, -- 1=coins 2=gems
    price_amount    INTEGER NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE user_items (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id     UUID NOT NULL REFERENCES store_items(id) ON DELETE CASCADE,
    acquired_at BIGINT NOT NULL,
    UNIQUE (user_id, item_id)
);

CREATE INDEX idx_store_items_type ON store_items(item_type);
CREATE UNIQUE INDEX uq_store_items_type_value ON store_items(item_type, item_value);
CREATE INDEX idx_user_items_user_id ON user_items(user_id);

INSERT INTO store_items (id, item_type, item_value, item_key, asset_url, price_currency, price_amount, is_active)
VALUES
    (uuidv7(), 1, 1, 'theme_default', '/assets/themes/1.png', 1, 0, true),
    (uuidv7(), 2, 1, 'xo_classic',    '/assets/xos/1.png',    1, 0, true),
    (uuidv7(), 3, 1, 'avatar_default','/assets/avatars/1.png',1, 0, true);
