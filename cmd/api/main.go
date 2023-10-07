package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/adarsh-shahi/gotube-api/internals/db"
)

type appConfig struct {
	port              string
	jsonSizeLimit     int
	dbConnectionCreds string
	DB *db.PostgreDB
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	
	app := appConfig{
		port:          ":8001",
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
	app.DB = &db.PostgreDB{PDB: conn, DbTimeout: 3 * time.Second}
	defer app.DB.Connection().Close()

	log.Println("Listening on port ", app.port)
	err = http.ListenAndServe(app.port, app.routes())
	if err != nil {
		log.Println(err)
	}
}
