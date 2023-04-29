package driver

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

// OpenDb opens connection to database
func OpenDb() (*sql.DB, error) {
	//
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	fmt.Println(os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
