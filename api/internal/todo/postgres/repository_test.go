package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"todo/api/internal/platform"
	"todo/api/internal/todo/domain"
	"todo/api/internal/todo/postgres"
)

func TestPostgresLifecycle(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set; PostgreSQL integration test skipped")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(context.Background())
	schema := fmt.Sprintf("todo_test_%d_%d", os.Getpid(), time.Now().UnixNano())
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := admin.Exec(cleanupCtx, "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	}()
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		parsed, err := url.Parse(databaseURL)
		if err != nil {
			t.Fatal(err)
		}
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		databaseURL = parsed.String()
	} else {
		databaseURL += " search_path=" + schema
	}
	pool, err := platform.OpenDatabase(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := postgres.New(pool)
	todos, err := repo.List(ctx)
	if err != nil || len(todos) != 0 {
		t.Fatalf("initial list = %v, %v", todos, err)
	}
	first, err := repo.Add(ctx, "first")
	if err != nil || first.ID <= 0 || first.Title != "first" || first.Done || first.CreatedAt.IsZero() {
		t.Fatalf("add = %+v, %v", first, err)
	}
	// A second startup must preserve existing data and tolerate the existing schema.
	secondPool, err := platform.OpenDatabase(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer secondPool.Close()
	repo = postgres.New(secondPool)
	second, err := repo.Add(ctx, "second")
	if err != nil {
		t.Fatal(err)
	}
	todos, err = repo.List(ctx)
	if err != nil || len(todos) != 2 || todos[0].ID != second.ID || todos[1].ID != first.ID {
		t.Fatalf("list = %+v, %v", todos, err)
	}
	for _, done := range []bool{true, false} {
		updated, err := repo.SetDone(ctx, first.ID, done)
		if err != nil || updated.Done != done || updated.Title != first.Title || !updated.CreatedAt.Equal(first.CreatedAt) {
			t.Fatalf("toggle = %+v, %v", updated, err)
		}
	}
	if err := repo.Delete(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.SetDone(ctx, first.ID, true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing toggle = %v", err)
	}
	if err := repo.Delete(ctx, first.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing delete = %v", err)
	}
	todos, err = repo.List(ctx)
	if err != nil || len(todos) != 1 || todos[0].ID != second.ID {
		t.Fatalf("after delete = %+v, %v", todos, err)
	}
	for _, title := range []string{"", " padded", "trailing ", strings.Repeat("界", 201)} {
		if _, err := repo.Add(ctx, title); err == nil {
			t.Fatalf("database accepted invalid title %q", title)
		}
	}
}
