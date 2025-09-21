-- for gen_random_uuid() , enable pgcrypto extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- User Table
CREATE TABLE IF NOT EXISTS users (
    id  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username text UNIQUE NOT NULL,
    password_hash text NOT NULL,
    email text UNIQUE ,
    used_storage bigint DEFAULT 0,
    quota bigint DEFAULT 10485760,   --default 10MB
    created_at timestamptz DEFAULT now()
);

-- Files Table (deduplicated objects)
CREATE TABLE IF NOT EXISTS files (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    hash_text text UNIQUE NOT NULL,   --sha256 hex string
    object_name text NOT NULL,
    size bigin NOT NULL,
    mime_type text,
    ref_count int DEFAULT 1,
    created_at timestamptz DEFAULT now()
);

-- User Files Table
CREATE TABLE user_files (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user(id) ON DELETE CASCADE,
    file_id uuid NOT NULL REFERENCES file(id) ON DELETE CASCADE,
    is_owner boolean DEFAULT false,
    visibility text DEFAULT 'private', 
    downloads bigint DEFAULT 0,
    created_at timestamptz DEFAULT now(),
    UNIQUE(user_id, file_id)
);

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
