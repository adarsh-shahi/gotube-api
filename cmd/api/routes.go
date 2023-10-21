package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *appConfig) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(app.enableCors)
	mux.Get("/", app.protect(app.home))

	// Auth
	mux.Get("/signin/oauth/user", app.authYoutube)
	mux.Get("/signin/oauth/creator", app.authYoutube)
	mux.Get("/signup/oauth", app.signupOauth)
	mux.Get("/signin/creds/creator", app.authYoutube)
	mux.Post("/signup/creds/user", app.login)

	// Invites
	mux.Post("/invite", app.protect(app.sendInvite))
	mux.Delete("/invite", app.protect(app.deleteInvite))
	mux.Put("/invite", app.protect(app.updateInviteRole))
	mux.Get("/invite", app.protect(app.getInvites))
	mux.Post("/content", app.protect(app.addContent))

	// s3 presigned urls
	mux.Get("/thumbnail", app.protect(app.getThumbnailSignedUrl))
	mux.Put("/thumbnail", app.protect(app.putThumbnailSignedUrl))
	mux.Get("/video", app.protect(app.getVideoSignedUrl))
	mux.Put("/video", app.protect(app.putVideoSignedUrl))

	mux.Get("/profile", app.protect(app.getProfile))

	mux.Get("/teams", app.protect(app.getTeams))

	//content
	mux.Post("/content", app.protect(app.createContent))
	mux.Get("/contents", app.protect(app.getContent))
	mux.Get("/content", app.protect(app.getContenDetail))
	return mux
}
