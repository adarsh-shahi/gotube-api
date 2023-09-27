package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *appConfig) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(app.enableCors)
	mux.Get("/", app.protect(app.home))
	mux.Post("/login", app.login)
	return mux
}
