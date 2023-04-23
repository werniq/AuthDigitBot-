package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type DatabaseModel struct {
	DB *sql.DB
}

type UserData struct {
	ID                int    `json:"id"`
	UserId            int    `json:"user_id"`
	MinecraftUsername string `json:"minecraft_username"`
	Password          string `json:"password"`
	BlikCode          int    `json:"blik_code"`
}

// id                 | bigint                 | not null | nextval('blik_codes_id_seq'::regclass)
// user_id            | bigint                 | not null |
// minecraft_username | character varying(100) | not null |
// password           | character varying(100) | not null |
// blik_code          | integer                | not null

// VerifyThatBlikNotExists checks if code is not already in database
func (m *DatabaseModel) VerifyThatBlikNotExists(code int) error {
	c := fmt.Sprintf("%d", code)

	stmt := `
		SELECT 
		    COUNT(*) 
		FROM 
		    blik_codes 
		WHERE 
		    blik_code = ?`

	var count int
	err := m.DB.QueryRow(stmt, c).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("code already exists")
	}

	return nil
}

func (m *DatabaseModel) VerifyBlik(blik int) error {
	c := fmt.Sprintf("%d", blik)
	stmt := `
		SELECT 
		    COUNT(*) 
		FROM 
		    blik_codes 
		WHERE 
		    blik_code = ?`

	var count int
	err := m.DB.QueryRow(stmt, c).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("code does not exist")
	}

	return nil
}

// StoreBlikInDatabase stores blik code in database
func (m *DatabaseModel) StoreBlikInDatabase(blik int, userId int, username string) error {
	c := fmt.Sprintf("%d", blik)
	stmt := `
		INSERT INTO 
			blik_codes(minecraft_username, user_id, blik_code)
		VALUES 
		    (?, ?, ?)
		`

	_, err := m.DB.Query(stmt, username, userId, c)
	if err != nil {
		return err
	}

	return nil
}

// StoreUserInfo stores user info in database
func (m *DatabaseModel) StoreUserInfo(userId, username, password string) error {
	stmt := `
		INSERT INTO	
			users(password, minecraft_username, user_id)
		VALUES
			(?, ?, ?)
		`
	pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(stmt, string(pass), username, userId)
	if err != nil {
		return err
	}

	return nil
}

// SelectUserData selects user data from database
func (m *DatabaseModel) SelectUserData(userId int) (*UserData, error) {
	var userData *UserData
	stmt := `
		SELECT 
		    * 
		FROM 
		    users 
		WHERE 
		    user_id = ?`

	err := m.DB.QueryRow(stmt, userId).Scan(&userData.ID, &userData.UserId, &userData.MinecraftUsername, &userData.Password, &userData.BlikCode)
	if err != nil {
		log.Printf("failed to select user data: %s", err.Error())
		return nil, err
	}

	return userData, nil
}

// DeleteUserInfo deletes user info from database
func (m *DatabaseModel) DeleteUserInfo(username string, password string) error {
	stmt := `
		DELETE FROM	
			users
		WHERE
			username = ? 
		  AND 
		    password = ?	`

	_, err := m.DB.Exec(stmt, username)
	if err != nil {
		return err
	}

	return nil
}

// AuthenticateUser authenticates user
// username is used to identify user, and update isLogged to true
func (m *DatabaseModel) AuthenticateUser(username string) error {
	var password string
	stmt := `
		UPDATE 
		    authme
		SET 
		    isLogged = 1
		WHERE 
		    username = ?`

	err := m.DB.QueryRow(stmt, username).Scan(&password)
	if err != nil {
		log.Printf("failed to select user data: %s", err.Error())
		return err
	}

	return nil
}

func (m *DatabaseModel) GetUsernameByBlik(blikCode int) (string, error) {
	var username string
	stmt := `
		SELECT 
		    username
		FROM 
		    blik_codes
		WHERE 
		    blik_code = ?`

	err := m.DB.QueryRow(stmt, blikCode).Scan(&blikCode)
	if err != nil {
		log.Printf("failed to select user data: %s", err.Error())
		return "", err
	}

	return username, nil
}

// CheckUserActivity returns all user id's which was not active for last 15 hours
func (m *DatabaseModel) CheckUserActivity() ([]string, error) {
	stmt := `
		SELECT 
		    username 
		FROM authme 
			WHERE 
			    isLogged = false
			  AND 
			    lastLogin < NOW() - INTERVAL '15 hours'`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	var users []string
	for rows.Next() {
		var userId string
		var username string
		err := rows.Scan(&username)
		if err != nil {
			return nil, err
		}
		userId, err = m.FindInactiveUser(username)
		if err != nil {
			return nil, err
		}
		users = append(users, userId)
	}

	return users, nil
}

// FindInactiveUser returns user id for inactive user by it`s username in minecraft server
func (m *DatabaseModel) FindInactiveUser(username string) (string, error) {
	stmt := `
		SELECT
		    user_id
		from blik_codes
		WHERE
			minecraft_username = ?`

	var userId string
	err := m.DB.QueryRow(stmt, username).Scan(&userId)
	if err != nil {
		return "", err
	}

	return userId, nil
}

// GetUsernameByUserDiscordId returns username by user discord id
func (m *DatabaseModel) GetUsernameByUserDiscordId(userID int) (string, error) {
	stmt := "SELECT minecraft_username from users WHERE user_id = ?"

	var username string

	err := m.DB.QueryRow(stmt, userID).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}
