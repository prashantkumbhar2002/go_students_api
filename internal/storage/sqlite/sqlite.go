package sqlite

import (
	"database/sql"
	"log/slog"

	"github.com/prashantkumbhar2002/go_students_api/internal/config"
	_ "github.com/mattn/go-sqlite3" // We are using _ to import the sqlite3 driver (Why? Because we are not using the sqlite3 driver in this file,)
)

type Sqlite struct {
	Db *sql.DB
}

func NewSqlite(cfg *config.Config) (*Sqlite, error) {

	// Open the SQLite database
	db, err := sql.Open("sqlite3", cfg.StoragePath)

	if err != nil {
		slog.Error("Error opening SQLite database", "error", err)
		return nil, err
	}

	// Test the Database connection
	if err := db.Ping(); err != nil {
		slog.Error("Error pinging SQLite database", "error", err)
		return nil, err
	}

	// Create the students table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			age INTEGER NOT NULL,
			email TEXT NOT NULL
		)
	`)

	if err != nil {
		slog.Error("Error creating students table in SQLite database", "error", err)
		return nil, err
	}
	slog.Info("Students table created successfully in SQLite database")

	// Return the Sqlite struct
	return &Sqlite{Db: db}, nil
}

func (s *Sqlite) CreateStudent(name string, email string, age int) (int64, error) {

	// Prepare the SQL statement - why? Because it is more efficient to prepare the statement once and then execute it multiple times. and also helps to prevent SQL injection.
	stmt, err := s.Db.Prepare("INSERT INTO students (name, email, age) VALUES (?, ?, ?)") // ? is a placeholder for the values
	if err != nil {
		slog.Error("Error preparing SQL statement to create student", "error", err)
		return 0, err
	}
	defer stmt.Close()

	// Execute the SQL statement
	result, err := stmt.Exec(name, email, age)
	if err != nil {
		slog.Error("Error executing SQL statement to create student", "error", err)
		return 0, err
	}

	// Get the last inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("Error getting last inserted ID from the database", "error", err)
		return 0, err // returning 0 value bcz for int64 it is default value and if we return error here then it will be difficult to handle the error in the caller function.
	}

	// Return the last inserted ID
	slog.Info("Student created successfully in SQLite database", "id", id)
	return id, nil
}