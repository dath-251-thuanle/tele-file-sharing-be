--- migrations/001_create_users_table.sql

--- table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_user_id BIGINT NOT NULL UNIQUE,
    username varchar(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()   
);

--- FOR LOCAL TESTING PURPOSES ONLY
CREATE TABLE files(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name VARCHAR(1024) NOT NULL,
    size_bytes BIGINT NOT NULL,
    fpath VARCHAR(2048) NOT NULL,
    fstatus VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE shares(
    id SERIAL PRIMARY KEY,
    file_id INT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    hash_link VARCHAR(512) NOT NULL UNIQUE,
    password_hash VARCHAR(512),
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ
);

--- index
CREATE INDEX IF NOT EXISTS idx_users_telegram_user_id ON users(telegram_user_id);


