package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sakojpa/tasker/config"
	"github.com/sakojpa/tasker/pkg/api"
	"github.com/sakojpa/tasker/utils"
	"log"
	"net/http"
	"os"
)

// NewServer create server with config to handle queries
func NewServer(c *config.Config) *http.Server {
	router := newRouter(c)
	srv := &http.Server{
		ReadTimeout:  c.Server.ReadTimeout,
		WriteTimeout: c.Server.WriteTimeout,
		IdleTimeout:  c.Server.IdleTimeout,

		Addr:     utils.GetServerAddr(c),
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
	r.Get("/api/tasks", authConnect(api.GetAllTasksHandler, c))
	r.Handle("/api/task", authConnect(api.TaskRouterHandler, c))
	r.Post("/api/task/done", authConnect(api.DoneTaskHandler, c))
	r.Post("/api/signin", api.AuthHandler)
	r.Handle("/*", http.FileServer(http.Dir(c.Server.StaticDir)))
	return r
}

func authConnect(
	handler func(w http.ResponseWriter, r *http.Request, ctx context.Context), c *config.Config,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var jwt string
			cookie, err := r.Cookie("token")
			if err == nil {
				jwt = cookie.Value
			}
			_, err = api.TokenValidate(jwt)
			if err != nil {
				utils.SentErrorJson(w, err.Error(), http.StatusUnauthorized)
				return
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), c.DB.Timeout)
		defer cancel()
		handler(w, r, ctx)
	}
}
