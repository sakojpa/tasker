package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sakojpa/tasker/config"
	"github.com/sakojpa/tasker/pkg/api"
	"log"
	"net/http"
	"os"
)

// getServerAddr returns server address composed of host and port from configuration.
func getServerAddr(c *config.Config) string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// NewServer create server with config to handle queries
func NewServer(c *config.Config) *http.Server {
	router := newRouter(c)
	srv := &http.Server{
		ReadTimeout:  c.Server.ReadTimeout,
		WriteTimeout: c.Server.WriteTimeout,
		IdleTimeout:  c.Server.IdleTimeout,

		Addr:     getServerAddr(c),
		Handler:  router,
		ErrorLog: log.New(os.Stderr, "HTTP ", log.LstdFlags),
	}
	return srv
}

func newRouter(c *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		middleware.URLFormat,
	)
	r.Get("/api/nextdate", api.RepeatTaskHandler)
	r.Get("/api/tasks", api.AuthConnect(api.GetAllTasksHandler, c))
	r.Get("/api/task", api.AuthConnect(api.EditTaskHandler, c))
	r.Post("/api/task", api.AuthConnect(api.AddTaskHandler, c))
	r.Put("/api/task", api.AuthConnect(api.UpdateTaskHandler, c))
	r.Delete("/api/task", api.AuthConnect(api.DeleteTaskHandler, c))
	r.Post("/api/task/done", api.AuthConnect(api.DoneTaskHandler, c))
	r.Post("/api/signin", func(w http.ResponseWriter, r *http.Request) { api.AuthHandler(w, r, c) })
	r.Handle("/*", http.FileServer(http.Dir(c.Server.StaticDir)))
	return r
}
