# omi-pilot-todo-v2

omi pilot v2: a simple todo app (Next.js frontend, Go backend, PostgreSQL).

## Run

    docker compose up --build

Then open http://localhost:3000.

## Test

    cd api && go vet ./... && go test ./...
    cd web && npm ci && npm run lint && npm run build

`DATABASE_URL` enables the Postgres tests (CI provides it).
