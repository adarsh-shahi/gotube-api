package main

import "net/http"

type appConfig struct {
	port string
	jsonSizeLimit int
}

func main(){
	app := appConfig{
		port: ":8000",
		jsonSizeLimit: 1024 * 1024 - 100000,  // 1MB
	}

	http.ListenAndServe(app.port, app.routes())

}
