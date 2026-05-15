CREATE TABLE IF NOT EXISTS links (
    slug        VARCHAR(100) PRIMARY KEY,
    destination TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS clicks (
    id         BIGSERIAL PRIMARY KEY,
    slug       VARCHAR(100) NOT NULL REFERENCES links(slug) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_agent TEXT,
    referer    TEXT,
    ip_address VARCHAR(45)
);

CREATE INDEX IF NOT EXISTS clicks_slug_idx ON clicks(slug);
CREATE INDEX IF NOT EXISTS clicks_clicked_at_idx ON clicks(clicked_at DESC);
