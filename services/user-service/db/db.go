package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Connect(dbSource string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbSource)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to the database successfully")
	return db, nil
}
