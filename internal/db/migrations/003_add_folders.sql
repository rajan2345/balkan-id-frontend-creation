-- Folders table
CREATE TABLE folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    parent_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Files can now belong to folders
ALTER TABLE user_files
ADD COLUMN folder_id UUID REFERENCES folders(id) ON DELETE CASCADE;
