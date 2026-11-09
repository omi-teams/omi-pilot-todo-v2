package migrations

import _ "embed"

// Todos is the initial, idempotent schema applied at startup.
//
//go:embed 001_todos.sql
var Todos string
