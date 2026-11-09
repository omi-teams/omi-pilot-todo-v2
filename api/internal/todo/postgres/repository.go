package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"todo/api/internal/todo/domain"
)

type Repository struct{ pool *pgxpool.Pool }

var _ domain.Repository = (*Repository)(nil)

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) List(ctx context.Context) ([]domain.Todo, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, title, done, created_at FROM todos ORDER BY created_at DESC, id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	todos := []domain.Todo{}
	for rows.Next() {
		var todo domain.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Done, &todo.CreatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, rows.Err()
}

func (r *Repository) Add(ctx context.Context, title string) (domain.Todo, error) {
	var todo domain.Todo
	err := r.pool.QueryRow(ctx, "INSERT INTO todos (title) VALUES ($1) RETURNING id, title, done, created_at", title).Scan(&todo.ID, &todo.Title, &todo.Done, &todo.CreatedAt)
	return todo, err
}

func (r *Repository) SetDone(ctx context.Context, id int64, done bool) (domain.Todo, error) {
	var todo domain.Todo
	err := r.pool.QueryRow(ctx, "UPDATE todos SET done = $2 WHERE id = $1 RETURNING id, title, done, created_at", id, done).Scan(&todo.ID, &todo.Title, &todo.Done, &todo.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = domain.ErrNotFound
	}
	return todo, err
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
