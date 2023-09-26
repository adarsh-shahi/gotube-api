package main

import (
	"log"
	"net/http"
)

func (app *appConfig) home(w http.ResponseWriter, r *http.Request){


	user := ""
	app.DB.PDB.QueryRow("select password from users where id = 1;").Scan(&user)

	log.Println(user)
	data := struct {
		Name string `json:"name"`
		Age int `json:"age"`
	}{}
	err, statusCode := app.readJSON(w, r, &data)
	if err != nil {
		app.errorJSON(w, err, statusCode)
		return
	}
	app.writeJSON(w, http.StatusAccepted, data)
}

func (app *appConfig) login(w http.ResponseWriter, r *http.Request){

	creds := struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}{}
	err, statusCode := app.readJSON(w, r, &creds)
	if err != nil {
		app.errorJSON(w, err, statusCode)
	}


}
