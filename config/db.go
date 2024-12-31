package config

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
	// DB2 *gorm.DB
)

func ConnectDB() {
	dbHost := "172.16.6.31"
	dbPort := "1433"
	dbUser := "sa"
	dbPass := "pass,123"

	connectToDB := func(dbName string) *gorm.DB {
		connection := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;encrypt=disable", dbHost, dbUser, dbPass, dbPort, dbName)
		db, err := gorm.Open(sqlserver.Open(connection), &gorm.Config{SkipDefaultTransaction: true})

		if err != nil {
			panic(fmt.Sprintf("Failed to connect to %s database: %s", dbName, err))
		}

		log.Printf("Connected to %s database succesfully\n", dbName)
		return db
	}

	DB = connectToDB("Portal_HelpDesk_CS")
	// DB2 = connectToDB("Portal_EXT_CNAF_Mobile")
}
