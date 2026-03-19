package scripts

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Connection string to the default 'postgres' database
	dsn := "postgres://admin:admin@localhost:5432/postgres?sslmode=disable"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer db.Close()

	// Check connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v. Make sure PostgreSQL is running and credentials are correct.", err)
	}

	fmt.Println("Connected to PostgreSQL server.")

	// Create database
	dbName := "auth_db"
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		fmt.Printf("Note: %v (The database might already exist)\n", err)
	} else {
		fmt.Printf("Database '%s' created successfully!\n", dbName)
	}
}
