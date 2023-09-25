package main

import "net/http"

func (app *appConfig) home(w http.ResponseWriter, r *http.Request){
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
