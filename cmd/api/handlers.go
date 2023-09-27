package main

import (
	"log"
	"net/http"

	"github.com/adarsh-shahi/gotube-api/internals/db"
)

func (app *appConfig) home(w http.ResponseWriter, r *http.Request) {
	user := ""
	app.DB.PDB.QueryRow("select password from users where id = 1;").Scan(&user)

	log.Println(user)
	data := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{}
	err, statusCode := app.readJSON(w, r, &data)
	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}
	app.writeJSON(w, http.StatusAccepted, data)
}

func (app *appConfig) login(w http.ResponseWriter, r *http.Request) {
	userToFind := db.TGmailPassword{}
	err, statusCode := app.readJSON(w, r, &userToFind)
	if err != nil {
		app.errorJSON(w, err, statusCode)
	}
	user, err := app.DB.GetEmailPasswordUser(userToFind)
	if err != nil {
		app.errorJSON(w, err, http.StatusNotFound)
		return
	}

	tokenClaims := map[string]string{
		"email": user.Email,
		"Utype": "user",
	}

	token, err := app.generateToken(tokenClaims)
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
