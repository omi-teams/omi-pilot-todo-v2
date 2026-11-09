package domain

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

type Todo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"createdAt"`
}

var (
	ErrInvalidTitle = errors.New("title must contain 1-200 characters after trimming")
	ErrNotFound     = errors.New("todo not found")
)

func NormalizeTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if !utf8.ValidString(title) || utf8.RuneCountInString(title) < 1 || utf8.RuneCountInString(title) > 200 {
		return "", ErrInvalidTitle
	}
	return title, nil
}

type Repository interface {
	List(context.Context) ([]Todo, error)
	Add(context.Context, string) (Todo, error)
	SetDone(context.Context, int64, bool) (Todo, error)
	Delete(context.Context, int64) error
}
