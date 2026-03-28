CREATE TABLE IF NOT EXISTS qr_codes (
    id         BIGSERIAL PRIMARY KEY,
    user_id    TEXT        NOT NULL,
    qr_token   TEXT        NOT NULL UNIQUE,
    url        TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_qr_codes_qr_token ON qr_codes (qr_token);
CREATE INDEX IF NOT EXISTS idx_qr_codes_user_id  ON qr_codes (user_id);
