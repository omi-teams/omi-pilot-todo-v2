package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"todo/api/internal/platform"
	"todo/api/internal/todo/app"
	"todo/api/internal/todo/domain"
	"todo/api/internal/todo/httpapi"
	"todo/api/internal/todo/postgres"
)

func newMux(repo domain.Repository) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health)
	httpapi.Register(mux, app.New(repo))
	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	pool, err := platform.OpenDatabase(ctx, os.Getenv("DATABASE_URL"))
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{Addr: ":" + port, Handler: newMux(postgres.New(pool)), ReadHeaderTimeout: 5 * time.Second}
	if err := server.ListenAndServe(); err != nil {
		log.Print(err)
	}
}
