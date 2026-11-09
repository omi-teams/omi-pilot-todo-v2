package app_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"todo/api/internal/todo/app"
	"todo/api/internal/todo/domain"
	"todo/api/internal/todo/todotest"
)

func TestLifecycle(t *testing.T) {
	ctx := context.Background()
	repo := &todotest.Repository{}
	service := app.New(repo)
	empty, err := service.List(ctx)
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("empty list = %v, %v", empty, err)
	}
	first, err := service.Add(ctx, "  first \n")
	if err != nil || first.Title != "first" || first.ID <= 0 || first.Done || first.CreatedAt.IsZero() {
		t.Fatalf("add = %+v, %v", first, err)
	}
	second, err := service.Add(ctx, "second")
	if err != nil {
		t.Fatal(err)
	}
	todos, err := service.List(ctx)
	if err != nil || !reflect.DeepEqual(todos, []domain.Todo{second, first}) {
		t.Fatalf("list = %+v, %v", todos, err)
	}
	for _, done := range []bool{true, false} {
		updated, err := service.SetDone(ctx, first.ID, done)
		if err != nil || updated.Done != done || updated.Title != first.Title || !updated.CreatedAt.Equal(first.CreatedAt) {
			t.Fatalf("set done = %+v, %v", updated, err)
		}
	}
	if err := service.Delete(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetDone(ctx, first.ID, true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing update = %v", err)
	}
	if err := service.Delete(ctx, first.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing delete = %v", err)
	}
	todos, err = service.List(ctx)
	if err != nil || len(todos) != 1 || todos[0].ID != second.ID {
		t.Fatalf("remaining = %v, %v", todos, err)
	}
	if _, err := service.Add(ctx, " \n"); !errors.Is(err, domain.ErrInvalidTitle) {
		t.Fatalf("invalid title = %v", err)
	}
	if len(repo.Todos) != 1 {
		t.Fatal("invalid title reached repository")
	}
}

func TestListOrdering(t *testing.T) {
	now := time.Now()
	repo := &todotest.Repository{Todos: []domain.Todo{{ID: 1, CreatedAt: now}, {ID: 3, CreatedAt: now}, {ID: 2, CreatedAt: now.Add(time.Second)}}}
	todos, err := app.New(repo).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for i, id := range []int64{2, 3, 1} {
		if todos[i].ID != id {
			t.Fatalf("ordering = %v", todos)
		}
	}
}

func TestRepositoryErrors(t *testing.T) {
	failure := errors.New("repository unavailable")
	service := app.New(&todotest.Repository{Err: failure})
	ctx := context.Background()
	_, listErr := service.List(ctx)
	_, addErr := service.Add(ctx, "title")
	_, updateErr := service.SetDone(ctx, 1, true)
	deleteErr := service.Delete(ctx, 1)
	for _, err := range []error{listErr, addErr, updateErr, deleteErr} {
		if !errors.Is(err, failure) {
			t.Fatalf("error = %v", err)
		}
	}
}
