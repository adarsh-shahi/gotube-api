package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/adarsh-shahi/gotube-api/internals/db"
	"golang.org/x/crypto/bcrypt"
)

func (app *appConfig) home(w http.ResponseWriter, r *http.Request) {
	log.Println("in home")
	log.Println(parsedUserData)
	app.writeJSON(w, http.StatusAccepted, "i did it")
}

func (app *appConfig) login(w http.ResponseWriter, r *http.Request) {
	userToFind := db.TIdEmailPassword{}
	err, statusCode := app.readJSON(w, r, &userToFind)
	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}
	user, err := app.DB.GetIdEmailPasswordUser(userToFind)
	if err != nil {
		app.errorJSON(w, err, http.StatusNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userToFind.Password))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			app.errorJSON(w, errors.New("wrong password"), http.StatusBadRequest)
		default:
			app.errorJSON(w, err, http.StatusUnauthorized)
		}
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

func (app *appConfig) signup(w http.ResponseWriter, r *http.Request) {
	user := db.AddUser{}
	err, statusCode := app.readJSON(w, r, &user)
	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}
	log.Println(user)

	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
	}
	user.Password = string(bytes)
	app.DB.AddUser(user)
	if err != nil {
		app.errorJSON(w, err, statusCode)
	}

	response := jsonResponse{
		Error:   false,
		Message: "account created",
	}
	app.writeJSON(w, http.StatusCreated, response)
}

func (app *appConfig) sendInvite(w http.ResponseWriter, r *http.Request) {
	user := struct {
		Email string `json:"email"`
	}{}

	err, statusCode := app.readJSON(w, r, &user)
	if err != nil {
		app.errorJSON(w, err, statusCode)
	}
}
