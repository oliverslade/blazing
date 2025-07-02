# Blazing

A Blazingly fast, simple Go web app with example user journey.

## Setup

1. Copy `.env.example` to `.env` and set your `DATABASE_URL`.
2. Run the SQL in `database/creation.sql` to set up your database.
3. Build and run the server:

```sh
go run ./cmd/web/main.go
```

## Environment Variables
- `DATABASE_URL`: PostgreSQL connection string.

## HTTPS
- Expects `fullchain.pem` and `privkey.pem` in the root directory for TLS.

## Static & Templates
- Static files: `static/`
- Templates: `templates/`
