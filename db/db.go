package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	Db *sql.DB
}

func NewStore(dbName string) (Store, error) {
	Db, err := getConnection(dbName)
	if err != nil {
		return Store{}, err
	}

	if err := createMigrations(Db); err != nil {
		return Store{}, err
	}

	return Store{
		Db,
	}, nil
}

func getConnection(dbName string) (*sql.DB, error) {
	var (
		err error
		db  *sql.DB
	)

	// Init SQLite3 database
	db, err = sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %s", err)
	}

	log.Println("Connected Successfully to the Database")

	return db, nil
}

func createMigrations(db *sql.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		password VARCHAR(255) NOT NULL,
		username VARCHAR(64) NOT NULL UNIQUE,
		changed_password BOOLEAN DEFAULT FALSE
	);`

	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), 8)
	if err != nil {
		return err
	}

	stmt = `INSERT INTO users(id, password, username, changed_password) VALUES($1, $2, $3, $4)`

	db.Exec(
		stmt,
		"1",
		string(hashedPassword),
		"admin",
		false,
	)

	return nil
}
