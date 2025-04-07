package main

import (
	"database/sql"
	"log"
	"os"
	"sync"

	queries "github.com/adisuper94/orcidparser/generated"
)

var q *queries.Queries
var db *sql.DB
var mutex = &sync.Mutex{}

func InitDB() {
	mutex.Lock()
	if db != nil {
		return
	}
	db, err := sql.Open("sqlite", "./orcid.db")
	if err != nil {
		log.Fatalln(err)
	}
	sqlBytes, err := os.ReadFile("./schema.sql")
	if err != nil {
		log.Fatalln(err)
	}
	sqlString := string(sqlBytes)
	_, err = db.Exec(sqlString)
	if err != nil {
		log.Fatalln(err)
	}
	q = queries.New(db)
	mutex.Unlock()
}

func GetQueries() *queries.Queries {
	if q == nil {
		InitDB()
	}
	return q
}
