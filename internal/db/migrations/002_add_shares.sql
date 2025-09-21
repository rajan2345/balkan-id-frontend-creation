CREATE TABLE shares (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id uuid NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN NOT NULL DEFAULT false,
    shared_with TEXT, 
    downloads INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
