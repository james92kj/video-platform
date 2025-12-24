-- Enable UUID extension 
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table 
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- videos table 
CREATE TABLE IF NOT EXISTS videos (
    
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    file_size BIGINT,
    duration INTEGER,
    original_filename VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for faster queries 
CREATE INDEX IF NOT EXISTS idx_videos_user_id ON videos(user_id);
CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);

-- Insert a test user
INSERT INTO users(id, username, email, password_hash)
VALUES('550e8400-e29b-41d4-a716-446655440000', 'testuser', 'test@example.com', 'hash123')
ON CONFLICT(id) DO NOTHING; 

