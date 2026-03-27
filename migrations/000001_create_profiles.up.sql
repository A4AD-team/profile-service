CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS profiles (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID         NOT NULL UNIQUE,
    username      VARCHAR(50)  NOT NULL UNIQUE,
    email         VARCHAR(190) NOT NULL UNIQUE,
    full_name     VARCHAR(190),
    bio           TEXT,
    location      VARCHAR(100),
    is_public     BOOLEAN      NOT NULL DEFAULT TRUE,
    post_count    INT          NOT NULL DEFAULT 0,
    comment_count INT          NOT NULL DEFAULT 0,
    reputation    INT          NOT NULL DEFAULT 0,
    joined_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_profiles_user_id   ON profiles (user_id);
CREATE INDEX IF NOT EXISTS idx_profiles_username  ON profiles (username);
CREATE INDEX IF NOT EXISTS idx_profiles_reputation ON profiles (reputation DESC);
