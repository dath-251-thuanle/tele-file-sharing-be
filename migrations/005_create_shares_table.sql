CREATE TABLE IF NOT EXISTS shares (
    id SERIAL PRIMARY KEY,
    file_id INT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    owner_user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Khóa ngoại tới id trong bảng users gốc 
    hash VARCHAR(255) NOT NULL UNIQUE, -- Dùng để tạo link chia sẻ public
    require_password BOOLEAN,
    hash_password TEXT,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ, -- Thời gian hết hạn (nếu có)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shares_hash ON shares(hash);
CREATE INDEX IF NOT EXISTS idx_shares_owner_user_id ON shares(owner_user_id);