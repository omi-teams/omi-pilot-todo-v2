# omi-pilot-todo-v2

omi pilot v2: a simple todo app (Next.js frontend, Go backend, PostgreSQL).

## Run

    docker compose up --build

Then open http://localhost:3000.

## Test

    cd api && go vet ./... && go test ./...
    cd web && npm ci && npm run lint && npm run build

`DATABASE_URL` enables the Postgres tests (CI provides it).

## API

The Go server requires `DATABASE_URL` and listens on `PORT` (default `8080`).
At startup it applies the embedded `api/migrations/001_todos.sql` in a transaction;
restarting preserves existing todos. Docker Compose supplies the database URL.

- `GET /api/health` returns `{"status":"ok"}`.
- `GET /api/todos` returns an array, newest first (an empty list is `[]`).
- `POST /api/todos` with `{"title":"Buy milk"}` returns the new todo with status 201.
  Titles are trimmed and must contain 1–200 Unicode characters; invalid titles return 400.
- `PATCH /api/todos/{id}` with `{"done":true}` or `{"done":false}` returns the updated todo.
- `DELETE /api/todos/{id}` returns 204. Updating or deleting an unknown ID returns 404.

Todos have `id`, `title`, `done`, and `createdAt` (RFC 3339) fields.
Errors are JSON objects with an `error` message.

The PostgreSQL integration test creates and removes an isolated schema, so the test
role needs permission to create schemas. Without `DATABASE_URL` it is skipped;
domain, application, HTTP, and health tests still run without a database.
