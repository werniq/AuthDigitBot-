package models

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
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

type User struct {
	ID int `json:"id"`
	// DiscordUserId is a user id in the Discord
	DiscordUserId string `json:"discord_user_id"`

	// MinecraftUsername is a username in the game
	MinecraftUsername string `json:"minecraft_username"`

	// Password is a hashed password
	Password string `json:"password"`

	// TotalMessagesCount is a total amount of messages sent by the user
	TotalMessagesCount int `json:"messages_count"`

	// ToRewardMessageCount is amount of messages, on user's way to receive reward
	ToRewardMessageCount int `json:"to_reward_messages_count"`

	// MinutesInVoice is a total amount of minutes spent in the voice chat
	MinutesInVoice int `json:"minutes_in_voice"`

	// Role is a role in the game, and in the Discord
	Role string `json:"role"`

	// Level means "1 уровень", or "2 уровень", etc.
	Level string `json:"level"`

	// Experience is a total amount of experience points
	Experience int `json:"experience"`

	// PasswordChanged at is a time when the user changed his password
	// will be used to check if the user changed password in last hour
	PasswordChangedAt time.Time `json:"password_changed_at"`

	// LastLogin is a time when the user logged in the game
	LastLogin time.Time `json:"last_login"`

	// CreatedAt is a time when the user was created
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        int       `json:"id"`
	MessageId string    `json:"message_id"`
	UserId    string    `json:"user_id"`
	ChannelId string    `json:"channel_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Quizes struct is a model for quizes table
type Quizes struct {
	ID int `json:"id"`

	// AuthorId is id of quiz creator
	AuthorId string `json:"user_id"`

	// QuizTitle specifies the title of the quiz
	QuizTitle string `json:"quiz_title"`

	// QuizQuestion is a string of questions, separated by comma
	QuizQuestion string `json:"quiz_questions"`

	// QuizAnswers is a string of answers, separated by comma
	QuizAnswers string `json:"quiz_answers"`

	// QuizExperienceReward is a string of experience reward
	QuizExperienceReward string `json:"quiz_experience_reward"`

	// QuizServerCurrencyReward is a string of server currency reward
	QuizServerCurrencyReward string `json:"quiz_server_currency_reward"`

	// CreatedAt is a time when the quiz was created
	CreatedAt time.Time `json:"created_at"`
}

var (
	levelIncreaseReward = map[string]int{
		"Генерал":           200,
		"Главарь Банды":     200,
		"Судья":             500,
		"Предприниматель":   1000,
		"Админ":             1500,
		"Модератор":         850,
		"Рядовой":           10,
		"Сержант":           20,
		"Лейтенант":         30,
		"Майор":             40,
		"Полковник":         50,
		"Мелкий":            10,
		"Бандит":            20,
		"Киллер":            30,
		"Заместитель главы": 40,

		"Gold": 250,
		"Star": 550,
		"VIP":  9500,
		"OG":   2000,
	}

	dailyExperienceReward = map[string]int{
		"1 уровень": 5,
		"2 уровень": 10,
		"3 уровень": 15,
		"4 уровень": 20,
		"5 уровень": 25,
	}

	dailyMoneyReward = map[string]int{
		"1 уровень": 0,
		"2 уровень": 0,
		"3 уровень": 50,
		"4 уровень": 100,
		"5 уровень": 150,
	}
)

// AddExperience adds experience to user
// TODO: update table name
func (m *DatabaseModel) AddExperience(experienceInt int, userId string) error {
	// here should be database where we store experience
	stmt := `UPDATE users SET experience = experience + $1 WHERE username = $1`

	username, err := m.GetUsernameByDiscordId(userId)
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(stmt, experienceInt, username)
	if err != nil {
		return err
	}

	return nil
}

// GetUsernameByDiscordId returns username by user's Discord ID
func (m *DatabaseModel) GetUsernameByDiscordId(userId string) (string, error) {
	stmt := `
		SELECT 
		    username 
		FROM 
		    users 
		WHERE 
		    discord_user_id = $1
		    `
	var u string

	err := m.DB.QueryRow(stmt, userId).Scan(&u)
	if err != nil {
		return "", err
	}

	return u, nil
}

// DailyExperienceAddingProcedure adds experience to all users, based on their level
func (m *DatabaseModel) DailyExperienceAddingProcedure() error {
	stmt := `SELECT discord_user_id from users`

	row, err := m.DB.Query(stmt)
	if err != nil {
		return err
	}

	defer row.Close()

	var userIds []string
	for row.Next() {
		var userId string
		err := row.Scan(&userId)
		if err != nil {
			return err
		}
		userIds = append(userIds, userId)
	}

	for _, userId := range userIds {
		err = m.DailyAddingExperience(userId)
		if err != nil {
			return err
		}
	}

	return nil
}

// DailyAddingExperience adds experience to user
// TODO: update table name
func (m *DatabaseModel) DailyAddingExperience(userId string) error {
	username, err := m.GetUsernameByDiscordId(userId)

	if err != nil {
		return err
	}

	stmt := `
		UPDATE
		    experience
		SET
		    experience = experience + $1
		WHERE
		    username = $2

		`
	err = m.DB.QueryRow(stmt, dailyExperienceReward["1 уровень"], username).Scan()

	return nil
}

// DailyAddingServerCurrency adds server currency to user
// Todo: update table name
func (m *DatabaseModel) DailyAddingServerCurrency(username string, amount int) error {
	stmt := `
			UPDATE
				experience
			SET 
			    server_currency = server_currency + $1
			WHERE
			    username = $2
		`

	err := m.DB.QueryRow(stmt, amount, username).Scan()
	if err != nil {
		return err
	}

	return nil
}

// GetLevelByUserID returns level of the user based on his Discord ID
func (m *DatabaseModel) GetLevelByUserID(userId string) (string, error) {
	stmt := "SELECT level from users where discord_user_id = $1"

	var level string

	err := m.DB.QueryRow(stmt, userId).Scan(&level)
	if err != nil {
		return "", err
	}

	return level, nil
}

// GetLevelByTheirUsername returns level of the user based on his Minecraft username
func (m *DatabaseModel) GetLevelByTheirUsername(username string) (string, error) {
	stmt := "SELECT level from users where minecraft_username = $1"

	var level string

	err := m.DB.QueryRow(stmt, username).Scan(&level)
	if err != nil {
		return "", err
	}

	return level, nil
}

// ChangeRole is helper function which is used to change role of the user
// TODO: change table name
func (m *DatabaseModel) ChangeRole(userId string, role string) error {
	stmt := `
		UPDATE 
		    users 
		SET 
		    role = $1 
		WHERE 
		    discord_user_id = $2
		    `

	_, err := m.DB.Exec(stmt, role, userId)
	if err != nil {
		return err
	}

	return nil
}

// GetRoleByUsername is helper function which is used to retrieve role by username
// TODO: change table name
func (m *DatabaseModel) GetRoleByUsername(username string) (string, error) {
	stmt := `
		SELECT 
		    role 
		FROM 
		    users 
		WHERE 
		    username = $1
		    `
	var r string

	err := m.DB.QueryRow(stmt, username).Scan(&r)
	if err != nil {
		return "", err
	}

	return r, nil
}

// GetCurrencyAmountByRole is helper function which is used to retrieve amount of currency which should be send to user
func (m *DatabaseModel) GetCurrencyAmountByRole(role string) int {
	return dailyMoneyReward[role]
}

// GetExperienceAmountByRole is helper function which is used to retrieve amount of experience which should be send to user
func (m *DatabaseModel) GetExperienceAmountByRole(role string) int {
	return dailyExperienceReward[role]
}

// GetUserMessagesCount is helper function which is used to retrieve amount of messages sent by user
// Todo: update table name
func (m *DatabaseModel) GetUserMessagesCount(userId string) error {
	stmt := `
		SELECT 
		    to_reward_messages_count, role 
		FROM 
		    users 
		WHERE 
		    discord_user_id = $1
		    `

	var count int
	var role string

	err := m.DB.QueryRow(stmt, userId).Scan(&count, &role)
	if err != nil {
		return err
	}

	if count >= 100 {
		err = m.AddExperience(m.GetExperienceAmountByRole(role), userId)
		if err != nil {
			return err
		}
	}

	return nil
}

// UserMessagesLeftToReward returns amount of messages left to receive reward
// TOdo: update table name
func (m *DatabaseModel) UserMessagesLeftToReward(userId string) (int, error) {
	stmt := "SELECT to_reward_messages_count from users where discord_user_id = $1"

	var c int

	err := m.DB.QueryRow(stmt, userId).Scan(&c)
	if err != nil {
		return 101, err
	}
	return 100 - c, nil
}

// AddExperienceForHundredMessages is helper function which is used to add experience to user
// TODO: update table name
func (m *DatabaseModel) AddExperienceForHundredMessages(userId string) error {
	var level string
	var username string

	stmt := `
		UPDATE 
		    users 
		SET 
		    to_reward_messages_count = 0 
		WHERE 
		    discord_user_id = $1
		    `

	_, err := m.DB.Exec(stmt, userId)
	if err != nil {
		return err
	}

	level, err = m.GetLevelByUserID(userId)
	if err != nil {
		return err
	}

	stmt = "UPDATE experience SET experience = experience + $1 WHERE username = $2"

	username, err = m.GetUsernameByDiscordId(userId)
	if err != nil {
		return err
	}

	if dailyMoneyReward[level] != 0 {
		err = m.DailyAddingServerCurrency(username, dailyMoneyReward[level])
		if err != nil {
			return err
		}
	}

	return nil
}

// AddServerCurrency is helper function which is used to add server currency to user
// TODO: update table name
func (m *DatabaseModel) AddServerCurrency(userId string, role string) error {
	username, err := m.GetUsernameByDiscordId(userId)
	if err != nil {
		return err
	}

	stmt := "UPDATE currency SET currency = currency + $1 WHERE username = $2"

	err = m.DB.QueryRow(stmt, dailyMoneyReward[role], username).Scan()
	if err != nil {
		return err
	}

	return nil

}

// AddServerCurrencyByAmount is helper function which is used to add server currency to user
func (m *DatabaseModel) AddServerCurrencyByAmount(amount int, userId string) error {
	username, err := m.GetUsernameByDiscordId(userId)
	if err != nil {
		return err
	}

	stmt := "UPDATE currency SET currency = currency + $1 WHERE username = $2"

	err = m.DB.QueryRow(stmt, amount, username).Scan()
	if err != nil {
		return err
	}

	return nil

}

// UserLastMessageTimestamp returns timestamp of the last message sent by user
func (m *DatabaseModel) UserLastMessageTimestamp(userId string) (*Message, error) {
	stmt := "SELECT * FROM messages where user_id = $1 ORDER BY timestamp DESC LIMIT 1"

	message := &Message{}

	err := m.DB.QueryRow(stmt, userId).Scan(&message.ID, &message.MessageId, &message.ChannelId, &message.UserId, &message.Content, &message.CreatedAt)

	if err != nil {
		return nil, err
	} else if err == sql.ErrNoRows {
		return nil, err
	}

	return message, nil
}

// CREATE table messages(
//	id bigserial not null primary key,
//	message_id varchar(50) not null,
//  channel_id int not null,
//	user_id varchar(50) not null,
//	content text,
//	timestamp timestamp not null);

// SaveMessageToDatabase is helper function which is used to save message to database
// TODO: Create messages table
func (m *DatabaseModel) SaveMessageToDatabase(message *Message) error {
	stmt := "INSERT INTO messages(message_id, channel_id, user_id, content, timestamp) VALUES ($1, $2, $3, $4, $5)"
	_, err := m.DB.Exec(stmt, message.MessageId, message.ChannelId, message.UserId, message.Content, message.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (m *DatabaseModel) GetAllUserIds() ([]string, error) {
	stmt := "SELECT discord_user_id FROM users"

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIds []string

	for rows.Next() {
		var userId string
		err = rows.Scan(&userId)
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, userId)
	}

	return userIds, nil
}

// GetAllUsernameAndIpPairs returns all existing username to ip address pairs
func (m *DatabaseModel) GetAllUsernameAndIpPairs() (map[string]string, error) {
	stmt := "SELECT minecraft_username, ip FROM ip_addresses"

	res, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	stmt = "SELECT ip FROM authme WHERE minecraft_username = $1"

	usernameToAddr := make(map[string]string)

	for i := 0; res.Next(); i++ {
		var username string
		var ip string

		err = res.Scan(&username, &ip)
		if err != nil {
			return nil, err
		}

		usernameToAddr[username] = ip
	}

	return usernameToAddr, nil
}

// CheckUserIPAddressChanging all users that have changed their ip address
func (m *DatabaseModel) CheckUserIPAddressChanging() ([]string, error) {
	usernameToIps, err := m.GetAllUsernameAndIpPairs()
	if err != nil {
		return nil, err
	}

	stmt := "SELECT ip, username from authme"

	res, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	existingUsernameToIps := make(map[string]string)

	for res.Next() {
		var ip string
		var username string

		err = res.Scan(&ip, &username)
		if err != nil {
			return nil, err
		}
		existingUsernameToIps[username] = ip
	}

	var usernames []string
	username := m.CompareMappings(usernameToIps, existingUsernameToIps)
	if username == "" {
		return usernames, nil
	} else {
		usernames = append(usernames, username)

		delete(usernameToIps, username)
		delete(existingUsernameToIps, username)
	}
	return nil, nil
}

// CompareMappings compares two maps and returns true if they are equal
func (m *DatabaseModel) CompareMappings(firstMap, secondMap map[string]string) string {
	for key, val1 := range firstMap {
		val2, ok := secondMap[key]
		if !ok || val1 != val2 {
			return key
		}
	}
	return ""
}

// ConvertUsernamesToUserIds converts usernames to user ids
func (m *DatabaseModel) ConvertUsernamesToUserIds(usernames []string) ([]string, error) {
	stmt := `select discord_user_id from users where minecraft_username = $1`

	var userIds []string

	for _, username := range usernames {
		var userId string
		err := m.DB.QueryRow(stmt, username).Scan(&userId)
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, userId)
	}

	return userIds, nil
}

// CheckUserActivity returns all userIds which have not been active for at 1 day
func (m *DatabaseModel) CheckUserActivity() ([]string, error) {
	stmt := "SELECT user_id FROM authme WHERE lastLogged < $1"

	var userIds []string

	yesterday := time.Now().AddDate(0, 0, -1)

	row, err := m.DB.Query(stmt, yesterday)
	if err != nil {
		return nil, err
	}

	for row.Next() {
		var userId string
		err = row.Scan(&userId)
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, userId)
	}

	return userIds, nil
}

// GetLastLogginUserTime returns last login time of the user
func (m *DatabaseModel) GetLastLogginUserTime(userId string) (int64, error) {
	stmt := `SELECT last_login FROM users WHERE discord_user_id = $1`

	var time time.Time

	err := m.DB.QueryRow(stmt, userId).Scan(&time)
	if err != nil {
		return time.Unix(), err
	}

	return time.Unix(), nil
}

// StoreUserInfo stores user info in the database
func (m *DatabaseModel) StoreUserInfo(userId, username, password, role, level string) error {
	p, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}

	user := &User{
		DiscordUserId:        userId,
		MinecraftUsername:    username,
		Password:             string(p),
		TotalMessagesCount:   0,
		ToRewardMessageCount: 0,
		MinutesInVoice:       0,
		Role:                 role,
		Level:                level,
		PasswordChangedAt:    time.Now(),
		LastLogin:            time.Now(),
		CreatedAt:            time.Now(),
	}

	stmt := `
  		INSERT INTO 
			users
  		    (discord_user_id, minecraft_username, 
  		     password, messages_count, 
  		     minutes_in_voice, role, level, 
  		     password_changed_at, last_login, created_at)
		VALUES
		  	($1, $2, $3, 
		  	 $4, $5, $6, 
		  	 $7, $8, $9, 
		  	 $10);
		`

	_, err = m.DB.Exec(stmt,
		user.DiscordUserId,
		user.MinecraftUsername,
		user.Password,
		user.TotalMessagesCount,
		user.MinutesInVoice,
		user.Role,
		user.Level,
		user.PasswordChangedAt,
		user.LastLogin,
		user.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

// CreateQuiz creates a quiz
func (m *DatabaseModel) CreateQuiz(quiz *Quizes) error {
	stmt := `INSERT INTO 
    			quizes( 
    	       			quiz_questions, quiz_correct_answers, 
    			        quiz_author, created_at)
    		 VALUES 
    		    ($1, $2, 
    		     $3, $4);`

	_, err := m.DB.Exec(stmt, quiz.QuizExperienceReward, quiz.QuizAnswers, quiz.AuthorId, quiz.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

// GetUserDataFromDatabaseByUserId returns user data from database by user id
func (m *DatabaseModel) GetUserDataFromDatabaseByUserId(userId string) (*User, error) {
	stmt := "SELECT * FROM users WHERE discord_user_id = $1"

	user := &User{}
	err := m.DB.QueryRow(stmt, userId).Scan(&user.DiscordUserId, &user.MinecraftUsername, &user.Password, &user.TotalMessagesCount, &user.ToRewardMessageCount, &user.MinutesInVoice, &user.Role, &user.Level, &user.PasswordChangedAt, &user.LastLogin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ComparePasswords compares hashed password with password
func (m *DatabaseModel) ComparePasswords(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// DeleteUserInfo deletes user info from the database
func (m *DatabaseModel) DeleteUserInfo(userId string) error {
	stmt := "DELETE FROM users WHERE discord_user_id = $1"

	_, err := m.DB.Exec(stmt, userId)
	if err != nil {
		return err
	}

	return nil
}

// GetLatestQuizes returns quizes which were created after the yesterday
func (m *DatabaseModel) GetLatestQuizes() ([]*Quizes, error) {
	stmt := "SELECT * from quizes where created_at > $1"

	var quizes []*Quizes
	res, err := m.DB.Query(stmt, time.Now().AddDate(0, 0, 0))
	if err != nil {
		return nil, err
	}

	defer res.Close()

	for res.Next() {
		quiz := &Quizes{}
		err := res.Scan(&quiz.ID, &quiz.QuizExperienceReward, &quiz.QuizAnswers, &quiz.AuthorId, &quiz.CreatedAt)
		if err != nil {
			return nil, err
		}
		quizes = append(quizes, quiz)
	}

	return quizes, nil
}

// CheckIfMessageIsAnswer checks if message is an answer to the quiz
func (m *DatabaseModel) CheckIfMessageIsAnswer(message string, userId string) (string, error) {
	stmt := "SELECT quiz_title, quiz_experience_reward, quiz_server_currency_reward FROM quizes WHERE quiz_correct_answers = $1 AND quiz_created_at > $2"

	var title string
	var experienceReward string
	var serverCurrencyReward string

	err := m.DB.QueryRow(stmt, message, time.Now().AddDate(0, 0, -1)).Scan(&title, &experienceReward, &serverCurrencyReward)
	if err != nil {
		return "", err
	}

	exp, err := strconv.Atoi(experienceReward)
	if err != nil {
		return "", err
	}

	serverCurrency, err := strconv.Atoi(serverCurrencyReward)
	if err != nil {
		return "", err
	}

	err = m.AddExperience(exp, userId)
	if err != nil {
		return "", err
	}

	err = m.AddServerCurrencyByAmount(serverCurrency, userId)
	if err != nil {
		return "", err
	}

	return title, nil
}

// SelectUsersWithHundredMessages returns users with more than 100 messages
func (m *DatabaseModel) SelectUsersWithHundredMessages() ([]*User, error) {
	stmt := "SELECT * FROM users where messages_count > 100"

	var users []*User

	res, err := m.DB.Query(stmt)

	if err != nil {
		return nil, err
	}

	for res.Next() {
		user := &User{}
		err := res.Scan(&user.DiscordUserId, &user.MinecraftUsername, &user.Password, &user.TotalMessagesCount, &user.ToRewardMessageCount, &user.MinutesInVoice, &user.Role, &user.Level, &user.PasswordChangedAt, &user.LastLogin, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (m *DatabaseModel) ChargeUsersWithMoney(users []*User) error {
	for i := 0; i < len(users); i++ {
		err := m.AddExperienceForHundredMessages(users[i].DiscordUserId)
		if err != nil {
			return err
		}
	}
	return nil
}

// StoreSecondsInVoiceChar stores seconds in voice chat for user based on his userId
func (m *DatabaseModel) StoreSecondsInVoiceChar(userId string, seconds int) error {
	stmt := "UPDATE users set minutes_in_voice = minutes_in_voice + $1 where discord_user_id = $2"

	_, err := m.DB.Exec(stmt, seconds, userId)
	if err != nil {
		return err
	}

	return nil
}

func (m *DatabaseModel) Config() (any, error) {
	var err error
	_, err = m.DB.Exec(`DROP TABLE IF EXISTS users;`)
	_, err = m.DB.Exec(`DROP TABLE IF EXISTS messages;`)
	_, err = m.DB.Exec(`DROP TABLE IF EXISTS quizes;`)
	_, err = m.DB.Exec(`DROP TABLE IF EXISTS ip_addresses;`)
	_, err = m.DB.Exec(
		`
		CREATE TABLE IF NOT EXISTS users(
				id BIGSERIAL PRIMARY KEY,  
				discord_user_id varchar(100) NOT NULL,  
				minecraft_username varchar(100) NOT NULL,  
				password varchar(100) NOT NULL,  
				messages_count INTEGER NOT NULL,  
				minutes_in_voice INTEGER NOT NULL,  
				experience INTEGER NOT NULL,
				role varchar(20) NOT NULL,  
				level varchar(10) NOT NULL,  
				password_changed_at TIMESTAMP NOT NULL,  
				last_login TIMESTAMP NOT NULL,  
				created_at TIMESTAMP NOT NULL
		);
     		`)
	if err != nil {
		return nil, err
	}
	_, err = m.DB.Exec(`
		CREATE TABLE IF NOT EXISTS ip_addresses (
		id SERIAL PRIMARY KEY,
		minecraft_username VARCHAR(255) NOT NULL,
		ip VARCHAR(39) NOT NULL,
		discord_user_ip VARCHAR(39)
	);
`)
	_, err = m.DB.Exec(`
			CREATE TABLE if not exists quizes (
		    	ID SERIAL PRIMARY KEY,
		    	author_id VARCHAR(255) NOT NULL,
		    	quiz_title VARCHAR(255) NOT NULL,
		    	quiz_question TEXT NOT NULL,
		    	quiz_correct_answers TEXT NOT NULL,
		    	quiz_experience_reward VARCHAR(255) NOT NULL,
		    	quiz_server_currency_reward VARCHAR(255) NOT NULL,
		    	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
		`)
	_, err = m.DB.Exec(`
		CREATE table messages(
			id bigserial not null primary key,
			message_id varchar(50) not null,	
			channel_id varchar(50) not null,
			user_id varchar(50) not null,
			content text,
			timestamp timestamp not null);`)
	res, err := m.DB.Query("SELECT ip, username from authme")
	if err != nil {
		return nil, err
	}
	for res.Next() {
		var ip string
		var username string
		err = res.Scan(&ip, &username)
		if err != nil {
			return nil, err
		}
		_, err = m.DB.Exec("INSERT INTO ip_addresses(minecraft_username, ip) VALUES ($1, $2)", username, ip)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
