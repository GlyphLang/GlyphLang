-- Glyph Database Initialization Script

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create schemas
CREATE SCHEMA IF NOT EXISTS glyph;

-- Example tables for Glyph applications

-- Users table
CREATE TABLE IF NOT EXISTS glyph.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table
CREATE TABLE IF NOT EXISTS glyph.sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES glyph.users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Posts table (example for blog/content)
CREATE TABLE IF NOT EXISTS glyph.posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES glyph.users(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    content TEXT,
    published BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON glyph.users(email);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON glyph.sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON glyph.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON glyph.posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_published ON glyph.posts(published);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION glyph.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to tables
DROP TRIGGER IF EXISTS update_users_updated_at ON glyph.users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON glyph.users
    FOR EACH ROW
    EXECUTE FUNCTION glyph.update_updated_at_column();

DROP TRIGGER IF EXISTS update_posts_updated_at ON glyph.posts;
CREATE TRIGGER update_posts_updated_at
    BEFORE UPDATE ON glyph.posts
    FOR EACH ROW
    EXECUTE FUNCTION glyph.update_updated_at_column();

-- Insert sample data
INSERT INTO glyph.users (email, name, password_hash) VALUES
    ('admin@example.com', 'Admin User', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('user@example.com', 'Regular User', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy')
ON CONFLICT (email) DO NOTHING;

-- Grant permissions
GRANT USAGE ON SCHEMA glyph TO glyph_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA glyph TO glyph_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA glyph TO glyph_user;
