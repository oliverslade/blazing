# Blazing

**Zero friction group chat for teams. One binary, zero dependencies.**

Blazing is an open source, self-hosted group chat server that compiles to a single small static binary. No external services, databases, or container runtimes required. Drop it on any cheap VPS, point a domain at it, and your team instantly has a private chat app with rooms, invites, and real time messaging.

## Why Blazing?

- **Zero Friction**: Single executable - no Docker, no databases to manage, no configuration hell
- **Small Footprint**: Runs on the cheapest VPS or Fly.io micro-VM (~$5/month)
- **Truly Private**: Self-hosted, GitHub OAuth only, no external services tracking your conversations
- **Instant Deploy**: `wget`, `chmod +x`, run - your chat server is live
- **Simple**: One audited Go codebase, easy to fork, extend, or embed in other tools
- **Frontend**: Server-rendered HTML + HTMX, no frontend build pipeline needed

## Quick Start

### Production Deployment

```bash
# Download and run
wget https://github.com/you/blazing/releases/latest/download/blazing
chmod +x blazing
./blazing
```

### Local Development

**Prerequisites:**

- Go 1.24 or later
- A GitHub account

**1. Create a GitHub OAuth App**

Go to https://github.com/settings/developers and create a new OAuth app:

- **Application name**: `Blazing Chat (Development)`
- **Homepage URL**: `http://localhost:8080`
- **Authorization callback URL**: `http://localhost:8080/auth/github/callback`

**2. Set up environment**

```bash
# Create .env file with your GitHub OAuth credentials
cat > .env << 'EOF'
GITHUB_CLIENT_ID=your_client_id_here
GITHUB_CLIENT_SECRET=your_client_secret_here
SESSION_SECRET=your-secret-key-at-least-32-characters-long
PORT=8080
DB_PATH=./blazing.db
GITHUB_REDIRECT_URL=http://localhost:8080/auth/github/callback
EOF

# Generate a secure session secret
SESSION_SECRET=$(openssl rand -base64 32)
sed -i "s/your-secret-key-at-least-32-characters-long/$SESSION_SECRET/" .env

# Start the server
export $(grep -v '^#' .env | xargs)
make run
```

Visit http://localhost:8080 and authenticate with GitHub!

## How It Works

**Core User Journey:**

1. **First visit**: "Sign in with GitHub" button
2. **Dashboard**: Create rooms or join existing ones
3. **Chat**: Real time messaging with WebSocket auto-reconnect
4. **Invites**: Add teammates by GitHub username - they're instantly in

**Technical Architecture:**

```
Browser ──HTTP──> chi router ──> SQLite ──> WebSocket hub
        └─HTMX────┘              ↑
                            embedded DB
```

Everything lives in one binary:

- **Database**: Embedded SQLite with WAL mode for concurrency
- **Auth**: GitHub OAuth with signed HTTP-only cookies
- **Real-time**: WebSocket fan out per room with automatic reconnect
- **UI**: Server-rendered HTML templates enhanced with HTMX
- **Static Assets**: CSS and templates compiled into binary via Go embed

## Technology Stack

- **Runtime**: Go 1.24 (single static binary)
- **Database**: SQLite with automatic migrations
- **HTTP**: chi router with middleware (logging, recovery, timeouts)
- **WebSockets**: nhooyr.io/websocket for real-time messaging
- **Frontend**: html/template + HTMX (no build step)
- **Auth**: GitHub OAuth with golang.org/x/oauth2
- **Query Generation**: sqlc for type-safe database queries

## Configuration

Blazing configures itself via environment variables:

```bash
# Required
GITHUB_CLIENT_ID=your_client_id_here
GITHUB_CLIENT_SECRET=your_client_secret_here
SESSION_SECRET=your-secret-key-at-least-32-characters-long

# Optional
PORT=8080
DB_PATH=./blazing.db
GITHUB_REDIRECT_URL=http://localhost:8080/auth/github/callback
GO_ENV=development
```

**Generate a secure session secret:**

```bash
export SESSION_SECRET=$(openssl rand -base64 32)
```

## Development Commands

```bash
make check          # Run all quality checks (fmt, vet, test)
make test           # Run unit tests only
make generate       # Regenerate database code (after schema changes)
make build          # Build production binary
make clean          # Clean build artifacts
make run            # Start development server
make clean-db       # Clean database files
```

## Database Schema

```sql
users            (id, github_uid, login, avatar_url, created_at, updated_at)
rooms            (id, name, creator_id, created_at, updated_at)
room_memberships (room_id, user_id, joined_at) -- composite PK
messages         (id, room_id, user_id, body, created_at)
```

All tables include automatic timestamps and foreign key constraints for data integrity. Migrations are embedded in the binary from `internal/db/migrations/`.

## Deployment

**On any Linux VPS:**

```bash
# Upload binary
scp blazing user@server:/home/user/
ssh user@server

# Run with systemd (recommended)
sudo mv blazing /usr/local/bin/
sudo systemctl enable --now blazing

# Or run directly
mkdir -p /data
./blazing
```

**Resource Requirements:**

- **RAM**: 10-50MB (scales with concurrent users)
- **CPU**: Minimal (single core sufficient for 100+ users)
- **Storage**: SQLite database grows ~1KB per message
- **Network**: WebSocket connections (typically 1-2KB/s per active user)

## Troubleshooting

### Common Issues

**"configuration error: GITHUB_CLIENT_ID environment variable is required"**

- Make sure you've created a GitHub OAuth app and set the environment variables
- Check that you've exported the variables: `export $(grep -v '^#' .env | xargs)`

**"configuration error: SESSION_SECRET must be at least 32 characters"**

- Generate a proper session secret: `export SESSION_SECRET=$(openssl rand -base64 32)`

**Can't connect to localhost:8080**

- Check that the server started successfully
- Make sure no other service is using port 8080
- Try a different port: `export PORT=3000`

**GitHub OAuth callback error**

- Verify your GitHub OAuth app callback URL matches exactly: `http://localhost:8080/auth/github/callback`
- Check that GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET are correct

**Database errors**

- The SQLite database is created automatically
- If you see permission errors, check that the current directory is writable
- Try a different DB_PATH: `export DB_PATH=./test.db`

### Verify Setup

Test your environment:

```bash
# Check Go version
go version  # Should be 1.24+

# Check environment variables
echo $GITHUB_CLIENT_ID
echo $GITHUB_CLIENT_SECRET
echo $SESSION_SECRET

# Test build
make check

# Test server startup (Ctrl+C to stop)
make run
```

## Security & Operations

- **No passwords**: GitHub OAuth eliminates credential management
- **CSRF protection**: All state-changing endpoints protected
- **Rate limiting**: Message posting rate-limited per user
- **Auto-reconnect**: WebSocket clients reconnect on connection drops
- **Graceful shutdown**: SIGTERM handling with 30s drain period
- **Health checks**: Built-in endpoints for monitoring

## Contributing

Blazing embraces the Go philosophy: simple, readable, maintainable code with minimal dependencies. The entire codebase is designed to be understood and modified by small teams.

```bash
git clone https://github.com/you/blazing
cd blazing
go mod tidy
make run
```

## License

MIT - Build something awesome!
