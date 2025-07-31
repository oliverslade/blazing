-- Initial database schema for Blazing Chat

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    github_uid BIGINT NOT NULL UNIQUE,
    login VARCHAR(255) NOT NULL UNIQUE,
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE rooms (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    creator_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- (many-to-many relationship)
CREATE TABLE room_memberships (
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (room_id, user_id)
);

CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- TODO: Add performance indexes for queries
-- TODO: Add constraints and validation rules