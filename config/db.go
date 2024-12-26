package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/denisenkom/go-mssqldb"
)

var DB *sql.DB

func InitDB() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Retrieve database configuration
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Build connection string
	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&encrypt=true&TrustServerCertificate=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Open database connection
	var errDB error
	DB, errDB = sql.Open("sqlserver", connString)
	if errDB != nil {
		log.Fatalf("Failed to connect to database: %v", errDB)
	}

	// Test connection
	errDB = DB.Ping()
	if errDB != nil {
		log.Fatalf("Failed to ping database: %v", errDB)
	}

	log.Println("Connected to database!")
}
