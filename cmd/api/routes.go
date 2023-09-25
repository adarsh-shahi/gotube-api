package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *appConfig) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Get("/", app.home)
	return mux
}
