package main

import "net/http"

func (app *appConfig) home(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("this is a home page"))
}
