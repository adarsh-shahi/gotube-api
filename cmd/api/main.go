package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adarsh-shahi/gotube-api/internals/db"
)

type appConfig struct {
	port              string
	jsonSizeLimit     int
	dbConnectionCreds string
	DB *db.PostgreDB
}

func main() {
	app := appConfig{
		port:          ":8000",
		jsonSizeLimit: 1024*1024 , // 1MB
		dbConnectionCreds: fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5",
			"localhost",
			5431,
			"postgres",
			"shahi",
			"gotube",
		),
	}
	conn, err := app.connectDB()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to database....")
	app.DB = &db.PostgreDB{PDB: conn}
	defer app.DB.Connection().Close()

	log.Println("Listening on port ", app.port)
	http.ListenAndServe(app.port, app.routes())
}
