-- users table: store only bcrypt hashes
CREATE TABLE users (
  id         SERIAL PRIMARY KEY,
  username   TEXT UNIQUE NOT NULL,
  pw_hash    BYTEA NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- sessions table: keeps a short-lived, random token per login
CREATE TABLE sessions (
  id        TEXT PRIMARY KEY,            -- 32-byte hex
  user_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires   TIMESTAMPTZ NOT NULL
);
CREATE INDEX ON sessions (expires);
