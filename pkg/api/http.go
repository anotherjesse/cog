package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	e     *Engine
	queue chan string
}

func NewServer() Server {
	return Server{
		e: NewEngine(),
	}
}

func Serve(listen string) error {
	s := NewServer()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/v1/models/{userName}/{modelName}/versions/{versionId}", s.modelOpenAPISpec)
	r.Get("/v1/models/{userName}/{modelName}/versions", s.modelVersions)
	r.Post("/v1/predictions", s.predictAPI)
	r.Get("/v1/predictions/{id}", s.getPredictions)

	s.queue = make(chan string, 100)
	go s.e.Run(s.queue)

	return http.ListenAndServe(listen, r)
}
