package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
	_ "github.com/lib/pq"
)

// helper struct
type User struct {
	ID       int64     `field:"id"`
	Username string    `field:"username"`
	Email    string    `field:"email"`
	Birthday time.Time `field:"birthday"`
}

var DB *sql.DB

func InitializeDB(connectionString string) {
	var err error

	DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Unable to connect to the PostgreSQL database: %s", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Unable to ping the PostgreSQL database: %s", err)
	}
	log.Println("Successfully connected to the database")

	_, err = DB.ExecContext(context.Background(), `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(100),
		email VARCHAR(100),
		birthday DATE,
		password VARCHAR(100)
		avatar VARCHAR(256)
	)`)
	if err != nil {
		log.Fatalf("Error creating users table: %s", err)
	}

	fmt.Println("Table created successfully")
}

// this is for registering a user
// the password passed in is meant to be already hashed
func AddUserDB(username string, email string, password string, dob time.Time) {
	statement := "INSERT INTO users (name, email, birthday, password) VALUES ($1, $2, $3, $4)"
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	_, err := DB.ExecContext(ctx, statement, username, email, dob, password)
	if err != nil {
		log.Printf("Error adding a user to the database: %s", err)
	} else {
		log.Println("User added successfully!")
	}
}

// this is for /api/v1/auth/user endpoint
func QueryUserDB(id int) (User, error) {
	query := "SELECT name, email, birthday FROM users WHERE id = $1"

	var err error
	var UserDB User

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	row := DB.QueryRowContext(ctx, query, id)

	if err = row.Scan(&UserDB.Username, &UserDB.Email, &UserDB.Birthday); err != nil {
		if err == sql.ErrNoRows {
			return UserDB, fmt.Errorf("user with id %d not found", id)
		}

		return UserDB, fmt.Errorf("error scanning user data from db: %w", err) 
	}
	
	return UserDB, nil
}

func ComparePassUserDB(email string, password string) (bool, error) {
	query := "SELECT password FROM users WHERE email = $1"

	var err error
	var dbPassword string

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second * 10)
	defer cancel()

	row := DB.QueryRowContext(ctx, query, email)

	if err = row.Scan(&dbPassword); err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("user with email %s not found", email)
		}
		return false, fmt.Errorf("error scanning password from db: %w", err)
	}

	if dbPassword != password {
		return false, fmt.Errorf("invalid credentials")
	}

	return true, nil
}

func Close() {
	if DB != nil {
		DB.Close()
		log.Println("Connection to the database closed.")
	}
}
