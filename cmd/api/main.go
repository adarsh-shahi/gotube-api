package main

import "net/http"

type appConfig struct {
	port string
}

func main(){
	app := appConfig{
		port: ":8000",
	}

	http.ListenAndServe(app.port, app.routes())

}
