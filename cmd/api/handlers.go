package main

import (
	"log"
	"net/http"

	"github.com/adarsh-shahi/gotube-api/internals/db"
)

func (app *appConfig) home(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusAccepted, "i did it")
}

func (app *appConfig) login(w http.ResponseWriter, r *http.Request) {
	userToFind := db.TIdGmailPassword{}
	err, statusCode := app.readJSON(w, r, &userToFind)
	if err != nil {
		app.errorJSON(w, err, statusCode)
	}
	user, err := app.DB.GetIdEmailPasswordUser(userToFind)
	if err != nil {
		app.errorJSON(w, err, http.StatusNotFound)
		return
	}


	token, err := app.generateToken(*user)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	log.Println(user)
	jsonResponse := jsonResponse{
		Error:   false,
		Message: "loggedIn successfully",
		Data: struct {
			Token string `json:"token"`
		}{Token: token},
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse)
}


func (app *appConfig) sendInvite(w http.ResponseWriter, r *http.Request){
	user := struct{
		Email string `json:"email"`
	}{}



	err, statusCode := app.readJSON(w, r, &user)
	if err != nil {
		app.errorJSON(w, err, statusCode)
	}

	



}
