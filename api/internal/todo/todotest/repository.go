// Package todotest provides the in-memory repository used by application and HTTP tests.
package todotest

import (
	"context"
	"sync"
	"time"

	"todo/api/internal/todo/domain"
)

type Repository struct {
	mu     sync.Mutex
	Todos  []domain.Todo
	Err    error
	nextID int64
}

func (r *Repository) List(context.Context) ([]domain.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]domain.Todo{}, r.Todos...), r.Err
}
func (r *Repository) Add(_ context.Context, title string) (domain.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return domain.Todo{}, r.Err
	}
	r.nextID++
	todo := domain.Todo{ID: r.nextID, Title: title, CreatedAt: time.Now().UTC()}
	r.Todos = append(r.Todos, todo)
	return todo, nil
}
func (r *Repository) SetDone(_ context.Context, id int64, done bool) (domain.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return domain.Todo{}, r.Err
	}
	for i := range r.Todos {
		if r.Todos[i].ID == id {
			r.Todos[i].Done = done
			return r.Todos[i], nil
		}
	}
	return domain.Todo{}, domain.ErrNotFound
}
func (r *Repository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return r.Err
	}
	for i := range r.Todos {
		if r.Todos[i].ID == id {
			r.Todos = append(r.Todos[:i], r.Todos[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}
