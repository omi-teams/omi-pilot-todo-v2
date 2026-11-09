package httpapi_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"todo/api/internal/todo/app"
	"todo/api/internal/todo/domain"
	"todo/api/internal/todo/httpapi"
	"todo/api/internal/todo/domain/todotest"
)

func handler(repo *todotest.Repository) *http.ServeMux {
	mux := http.NewServeMux()
	httpapi.Register(mux, app.New(repo))
	return mux
}
func request(t *testing.T, mux http.Handler, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	if rec.Code != status {
		t.Fatalf("%s %s: status %d, want %d; body %s", method, path, rec.Code, status, rec.Body.String())
	}
	if status != http.StatusNoContent && rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content type = %q", rec.Header().Get("Content-Type"))
	}
	return rec
}
func decodeTodo(t *testing.T, rec *httptest.ResponseRecorder) domain.Todo {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 4 {
		t.Fatalf("JSON fields = %v", fields)
	}
	for _, field := range []string{"id", "title", "done", "createdAt"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("missing %s", field)
		}
	}
	var todo domain.Todo
	if err := json.Unmarshal(rec.Body.Bytes(), &todo); err != nil {
		t.Fatal(err)
	}
	if todo.ID <= 0 || todo.CreatedAt.IsZero() {
		t.Fatalf("todo = %+v", todo)
	}
	return todo
}
func TestHTTPLifecycle(t *testing.T) {
	mux := handler(&todotest.Repository{})
	rec := request(t, mux, "GET", "/api/todos", "", 200)
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("empty = %s", rec.Body.String())
	}
	first := decodeTodo(t, request(t, mux, "POST", "/api/todos", `{"title":"  first  "}`, 201))
	if first.Title != "first" || first.Done {
		t.Fatalf("created = %+v", first)
	}
	second := decodeTodo(t, request(t, mux, "POST", "/api/todos", `{"title":"second"}`, 201))
	rec = request(t, mux, "GET", "/api/todos", "", 200)
	var todos []domain.Todo
	if err := json.Unmarshal(rec.Body.Bytes(), &todos); err != nil {
		t.Fatal(err)
	}
	if len(todos) != 2 || todos[0].ID != second.ID || todos[1].ID != first.ID {
		t.Fatalf("list = %+v", todos)
	}
	path := fmt.Sprintf("/api/todos/%d", first.ID)
	for _, done := range []bool{true, false} {
		todo := decodeTodo(t, request(t, mux, "PATCH", path, fmt.Sprintf(`{"done":%v}`, done), 200))
		if todo.Done != done || todo.Title != first.Title || !todo.CreatedAt.Equal(first.CreatedAt) {
			t.Fatalf("updated = %+v", todo)
		}
	}
	rec = request(t, mux, "DELETE", path, "", 204)
	if rec.Body.Len() != 0 {
		t.Fatal("204 has body")
	}
	request(t, mux, "PATCH", path, `{"done":true}`, 404)
	request(t, mux, "DELETE", path, "", 404)
	rec = request(t, mux, "GET", "/api/todos", "", 200)
	if err := json.Unmarshal(rec.Body.Bytes(), &todos); err != nil {
		t.Fatal(err)
	}
	if len(todos) != 1 || todos[0].ID != second.ID {
		t.Fatalf("remaining = %+v", todos)
	}
}
func TestBadRequests(t *testing.T) {
	for _, tc := range []struct{ method, path, body string }{
		{"POST", "/api/todos", `{"title":""}`},
		{"POST", "/api/todos", `{"title":" \n\t "}`},
		{"POST", "/api/todos", `{"title":"` + strings.Repeat("界", 201) + `"}`},
		{"POST", "/api/todos", `{}`},
		{"POST", "/api/todos", `null`},
		{"POST", "/api/todos", `{"title":12}`},
		{"POST", "/api/todos", `{"title":"ok"} {}`},
		{"POST", "/api/todos", `{"title":"ok","other":1}`},
		{"POST", "/api/todos", `{`},
		{"POST", "/api/todos", `{"title":"` + strings.Repeat("a", 17000) + `"}`},
		{"PATCH", "/api/todos/1", `{}`},
		{"PATCH", "/api/todos/1", `{"done":null}`},
		{"PATCH", "/api/todos/1", `{"done":"true"}`},
		{"PATCH", "/api/todos/1", `{"done":true} {}`},
		{"PATCH", "/api/todos/nope", `{"done":true}`},
		{"PATCH", "/api/todos/0", `{"done":true}`},
		{"DELETE", "/api/todos/-1", ``},
		{"DELETE", "/api/todos/9223372036854775808", ``},
	} {
		t.Run(tc.method+tc.path+fmt.Sprint(len(tc.body)), func(t *testing.T) {
			repo := &todotest.Repository{}
			rec := request(t, handler(repo), tc.method, tc.path, tc.body, 400)
			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] == "" {
				t.Fatalf("error response = %s", rec.Body.String())
			}
			if len(repo.Todos) != 0 {
				t.Fatal("bad request changed repository")
			}
		})
	}
}
func TestUnicodeTitleLimit(t *testing.T) {
	title := strings.Repeat("界", 200)
	todo := decodeTodo(t, request(t, handler(&todotest.Repository{}), "POST", "/api/todos", `{"title":" `+title+` "}`, 201))
	if todo.Title != title {
		t.Fatal("unicode title changed")
	}
}
func TestHTTPRepositoryErrors(t *testing.T) {
	for _, tc := range []struct{ method, path, body string }{
		{"GET", "/api/todos", ""}, {"POST", "/api/todos", `{"title":"ok"}`},
		{"PATCH", "/api/todos/1", `{"done":true}`}, {"DELETE", "/api/todos/1", ""},
	} {
		rec := request(t, handler(&todotest.Repository{Err: errors.New("secret database details")}), tc.method, tc.path, tc.body, 500)
		if strings.TrimSpace(rec.Body.String()) != `{"error":"internal server error"}` {
			t.Fatalf("error = %s", rec.Body.String())
		}
	}
}
func TestUnknownIDs(t *testing.T) {
	mux := handler(&todotest.Repository{Todos: []domain.Todo{{ID: 1, Title: "keep", CreatedAt: time.Now()}}})
	for _, tc := range []struct{ method, body string }{{"PATCH", `{"done":true}`}, {"DELETE", ""}} {
		rec := request(t, mux, tc.method, "/api/todos/999", tc.body, 404)
		if strings.TrimSpace(rec.Body.String()) != `{"error":"todo not found"}` {
			t.Fatalf("error = %s", rec.Body.String())
		}
	}
}
