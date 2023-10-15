package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *appConfig) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(app.enableCors)
	mux.Get("/", app.protect(app.home))
	mux.Get("/signin/oauth/user", app.authYoutube)
	mux.Get("/signin/oauth/creator", app.authYoutube)
	mux.Get("/signup/oauth", app.signupOauth)
	mux.Get("/signin/creds/creator", app.authYoutube)
	mux.Post("/signup/creds/user", app.login)
	mux.Post("/invite", app.protect(app.sendInvite))
	mux.Delete("/invite", app.protect(app.deleteInvite))
	mux.Put("/invite", app.protect(app.updateInviteRole))
	mux.Get("/invite", app.protect(app.getInvites))
	// mux.Get("/content", app.protect(app.getContent))
	// mux.Put("/content", app.protect(app.updateContent))
	mux.Post("/content", app.protect(app.addContent))
	// mux.Delete("/content", app.protect(app.deleteContent))
	mux.Get("/geturl", app.protect(app.getSignedUrl))
	mux.Get("/puturl", app.protect(app.putSignedUrl))
	mux.Get("/profile", app.protect(app.getProfile))
	return mux
}
