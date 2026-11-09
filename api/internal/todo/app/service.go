package app

import (
	"context"
	"sort"

	"todo/api/internal/todo/domain"
)

type Service struct{ repo domain.Repository }

func New(repo domain.Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]domain.Todo, error) {
	todos, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	if todos == nil {
		todos = []domain.Todo{}
	}
	sort.Slice(todos, func(i, j int) bool {
		if todos[i].CreatedAt.Equal(todos[j].CreatedAt) {
			return todos[i].ID > todos[j].ID
		}
		return todos[i].CreatedAt.After(todos[j].CreatedAt)
	})
	return todos, nil
}

func (s *Service) Add(ctx context.Context, title string) (domain.Todo, error) {
	title, err := domain.NormalizeTitle(title)
	if err != nil {
		return domain.Todo{}, err
	}
	return s.repo.Add(ctx, title)
}

func (s *Service) SetDone(ctx context.Context, id int64, done bool) (domain.Todo, error) {
	return s.repo.SetDone(ctx, id, done)
}

func (s *Service) Delete(ctx context.Context, id int64) error { return s.repo.Delete(ctx, id) }
