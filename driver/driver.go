package driver

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

// OpenDb opens connection to database
func OpenDb() (*sql.DB, error) {
	//
	db, err := sql.Open("mysql", os.Getenv("DATABASE_DSN"))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INT NOT NULL AUTO_INCREMENT, username VARCHAR(255) NOT NULL, password VARCHAR(255) NOT NULL, blik INT NOT NULL, PRIMARY KEY (id)")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("TRUNCATE TABLE users")

	_, err = db.Exec("CREATE TABLE blik_codes IF NOT EXISTS (id INT NOT NULL AUTO_INCREMENT, minecraft_username VARCHAR(255) NOT NULL, user_id VARCHAR(255) NOT NULL, blik_code INT NOT NULL, PRIMARY KEY (id)")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("TRUNCATE TABLE blik_codes")

	return db, nil
}
