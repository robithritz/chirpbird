package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/robithritz/chirpbird/common/database"
	"github.com/robithritz/chirpbird/common/router"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Cannot load .env")
	}

	database.ConnectDB()
	defer database.DB.Close()

	router.StartServer()

}
