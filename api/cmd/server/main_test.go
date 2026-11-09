package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"todo/api/internal/todo/todotest"
)

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux(&todotest.Repository{}).ServeHTTP(rec, httptest.NewRequest("GET", "/api/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Body.String(); got != `{"status":"ok"}` {
		t.Fatalf("body = %q", got)
	}
}

func TestTodoRoutesWired(t *testing.T) {
	mux := newMux(&todotest.Repository{})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("POST", "/api/todos", strings.NewReader(`{"title":"wired"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("post status = %d: %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/todos", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"title":"wired"`) {
		t.Fatalf("get = %d: %s", rec.Code, rec.Body.String())
	}
}
