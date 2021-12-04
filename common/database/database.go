package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

var DB *pgxpool.Pool
var err error

func ConnectDB() {

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("SSL_MODE")

	connStr := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=" + sslMode
	DB, err = pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Error Connection : ", err.Error())
	}

	pingError := DB.Ping(context.Background())
	if pingError != nil {

		log.Fatal(pingError)
	}
	fmt.Println("Connected!")

}
