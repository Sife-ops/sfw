package main

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var Db *sqlx.DB

func init() {
	if _, e := os.Stat("./db.sqlite"); e != nil {
		log.Fatalf("%v", e)
	}
	db, e := sqlx.Open("sqlite3", "./db.sqlite")
	if e != nil {
		log.Fatalf("%v", e)
	}
	Db = db
}
