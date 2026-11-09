package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"todo/api/internal/todo/app"
	"todo/api/internal/todo/domain"
)

func Register(mux *http.ServeMux, service *app.Service) {
	mux.HandleFunc("GET /api/todos", func(w http.ResponseWriter, r *http.Request) {
		todos, err := service.List(r.Context())
		respond(w, http.StatusOK, todos, err)
	})
	mux.HandleFunc("POST /api/todos", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Title string `json:"title"`
		}
		if !decode(w, r, &body) {
			return
		}
		todo, err := service.Add(r.Context(), body.Title)
		respond(w, http.StatusCreated, todo, err)
	})
	mux.HandleFunc("PATCH /api/todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := todoID(w, r)
		if !ok {
			return
		}
		var body struct {
			Done *bool `json:"done"`
		}
		if !decode(w, r, &body) {
			return
		}
		if body.Done == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "done must be a boolean"})
			return
		}
		todo, err := service.SetDone(r.Context(), id, *body.Done)
		respond(w, http.StatusOK, todo, err)
	})
	mux.HandleFunc("DELETE /api/todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := todoID(w, r)
		if !ok {
			return
		}
		respond(w, http.StatusNoContent, nil, service.Delete(r.Context(), id))
	})
}

func decode(w http.ResponseWriter, r *http.Request, body any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(body); err == nil {
		var extra any
		if decoder.Decode(&extra) == io.EOF {
			return true
		}
	}
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
	return false
}

func todoID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid todo id"})
		return 0, false
	}
	return id, true
}

func respond(w http.ResponseWriter, status int, value any, err error) {
	if err != nil {
		message := "internal server error"
		status = http.StatusInternalServerError
		switch {
		case errors.Is(err, domain.ErrInvalidTitle):
			status, message = http.StatusBadRequest, domain.ErrInvalidTitle.Error()
		case errors.Is(err, domain.ErrNotFound):
			status, message = http.StatusNotFound, domain.ErrNotFound.Error()
		}
		writeJSON(w, status, map[string]string{"error": message})
		return
	}
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return
	}
	writeJSON(w, status, value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
