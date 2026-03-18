cpd-hub-backend

This repository uses Clean Architecture (Hexagonal) for a Go backend.

Folder layout (high level):

- cmd/server - application entrypoint
- internal/domain - core entities and repository interfaces
- internal/usecase - business logic and worker pool implementation
- internal/infrastructure - persistence implementations and configuration
- internal/delivery - HTTP transport layer (handlers, server)

Getting started:

1. go mod tidy
2. go run ./cmd/server

This is a scaffold — implement your entities, usecases and infra as needed.

## Local development with Postgres (optional)

A `docker-compose.yml` is included to run a local Postgres instance for development. The project will automatically use Postgres when the `DATABASE_URL` environment variable is set.

Quick start (uses Docker):

1. Start Postgres:

   docker compose up -d

2. The compose file mounts `internal/infrastructure/postgres/seed.sql` into the container so initial seed data is applied on first startup. If you need to re-seed after the first run, see the "Seeding manually" section below.

3. Run the server (example):

   DATABASE_URL=postgres://postgres:postgres@localhost:5432/cpdhub go run ./cmd/server


## Seeding the database

A seed SQL file is provided at `internal/infrastructure/postgres/seed.sql` with example problems, contests, info and activity records. By default the Docker Compose setup will run the seed file on container initialization. To run the seed manually you can use `psql` (make sure `DATABASE_URL` is set) or use the helper script `scripts/seed_db.sh`.

Manual seed with psql:

   psql "$DATABASE_URL" -f internal/infrastructure/postgres/seed.sql

Or using the helper script:

   chmod +x scripts/seed_db.sh
   ./scripts/seed_db.sh


## Environment

- SERVER_ADDR - address the Go server will bind to (default `:8080`).
- DATABASE_URL - Postgres DSN (optional). When set the server will connect to Postgres and use DB-backed repositories.
- JWT_SECRET - secret used to sign JWT tokens (default fallback present for local dev).


## Example API usage

Sign up (creates user and returns token):

curl -X POST http://localhost:8080/api/auth/signup -H 'Content-Type: application/json' -d '{"fullName":"Alice","email":"alice","password":"secret","confirmPassword":"secret"}'

Login:

curl -X POST http://localhost:8080/api/auth/login -H 'Content-Type: application/json' -d '{"email":"alice","password":"secret"}'

Use token for protected request (replace <token> below):

curl -H "Authorization: Bearer <token>" http://localhost:8080/api/info


## Notes

- The seed file intentionally avoids inserting plaintext passwords. Use the signup endpoint to create users (it stores bcrypt-hashed passwords).
- For production use you should manage migrations with a proper migration tool and use a secure JWT secret.
