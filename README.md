# Blazing

## Core Journey

- First-time visit = “Sign in with GitHub” button.
- Dashboard
  - No rooms yet = show “Create a room” form.
  - Rooms exist = list joined rooms (name, last-message timestamp).
- Room creation = creator supplies a name, is auto-joined.
- Chat view
  - Live message stream.
  - “Invite” modal: enter GitHub login = user instantly added.
- Invitee flow = invited user logs in, sees the new room in their list.

## Feature Set

- Auth GitHub: OAuth, signed HTTP-only session cookie
- Room mgmt: Create, list, membership join/leave
- Messaging: Send/receive text; last-50 message history on load
- Invites: Lookup/auto-create user by GitHub login; idempotent add
- Realtime: WebSocket fan-out per room; automatic reconnect
- UI: HTML templates rendered server-side; HTMX progressively enhances forms; zero custom frontend build pipeline

## Technology Stack

- Language: Go 1.24
- HTTP Router: chi
- DB Access: sqlc-generated queries (PostgreSQL)
- Realtime: nhooyr.io/websocket
- Templating: html/template + HTMX Server-rendered HTML, progressive enhancement
- Sessions: SCS (signed cookies) No server-side store
- OAuth: golang.org/x/oauth2 (GitHub endpoint)

## Architecture Snapshot

```
Browser ──HTTP──> chi handlers ──> services ──> sqlc/Postgres
└─WS room:n──> in-memory hub ──> broadcast to peers
```

All assets (templates, embedded CSS) are compiled into the binary via embed; deploy = copy one file + run under systemd.

## Data Model

- users (id, github_uid, login, avatar_url)
- rooms (id, name, creator_id)
- room_memberships (room_id, user_id) — composite PK
- messages (id, room_id, user_id, body, created_at)

## Non-Functional Targets

- Resilience: auto-reconnect after process restart (<5 s); systemd restart policy.
- Security: no password storage; all endpoints CSRF protected; rate limit message POST.
