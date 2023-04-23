package main

import (
	"fmt"
	"freelance-discord/driver"
	"freelance-discord/models"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/bot/msg"
	"github.com/Tnze/go-mc/bot/playerlist"
	"github.com/Tnze/go-mc/chat"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	//rudderfish.aternos.host:34500
	"syscall"
)

type Application struct {
	bot *discordgo.Session
	db  *models.DatabaseModel
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bott, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	db, err := driver.OpenDb()
	if err != nil {
		log.Fatal(err)
	}

	var app *Application
	app = &Application{
		bot: bott,
		db:  &models.DatabaseModel{DB: db},
	}

	bott.Identify.Intents = discordgo.IntentsGuildMessages
	bott.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentsGuildPresences

	bott.AddHandler(app.GenerateBlikCode)
	bott.AddHandler(app.Register)
	bott.AddHandler(app.Stop)
	bott.AddHandler(app.CheckUserActivity)

	go app.CheckAuthentication()

	err = bott.Open()
	if err != nil {
		fmt.Println("Error opening Discord Session, ", err)
	}
	fmt.Println("Bot is currently running. CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

var chatHandler *msg.Manager

func (app *Application) CheckAuthentication() {
	client := bot.NewClient()

	err := client.JoinServer(os.Getenv("MINECRAFT_SERVER_ADDRESS"))

	if err != nil {
		panic(err)
	}
	var player *basic.Player
	pL := playerlist.New(client)

	chatHandler = msg.New(client, player, pL, msg.EventsHandler{
		SystemChat:        nil,
		PlayerChatMessage: app.PlayerChatMessage,
		DisguisedChat:     nil,
	})

	err = chatHandler.SendMessage("Please enter your blik code")
	if err != nil {
		panic(err)
	}

}

func (app *Application) PlayerChatMessage(msg chat.Message, validated bool) error {
	content := msg.Text

	blikCode, err := strconv.Atoi(content)
	if err != nil {
		if err := chatHandler.SendMessage("Invalid blik code"); err != nil {
			panic(err)
		}
		return err
	}

	err = app.db.VerifyThatBlikNotExists(blikCode)
	if err != nil {
		if err := chatHandler.SendMessage("error verifying blik code"); err != nil {
			panic(err)
		}
		return err
	}

	var username string
	username, err = app.db.GetUsernameByBlik(blikCode)
	if err != nil {
		if err := chatHandler.SendMessage("error getting username by blik code"); err != nil {
			panic(err)
		}
		return err
	}

	err = app.db.AuthenticateUser(username)
	if err != nil {
		if err := chatHandler.SendMessage("error authenticating user"); err != nil {
			panic(err)
		}
		return err
	}

	return nil
}

// GenerateBlikCode generates blik code
func (app *Application) GenerateBlikCode(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if m.Content == "/blik" {
		// Set up a timer that runs the code generation logic every 30 seconds
		timer := time.NewTicker(30 * time.Second)
		defer timer.Stop()

		for {
			// Generate a new code
			num := rand.Intn(10000000) + 1

			// Check if the code already exists in the database
			err := app.db.VerifyThatBlikNotExists(num)

			if err == nil {
				// Store the code in the database
				userId, err := strconv.Atoi(m.Author.ID)
				if err != nil {
					log.Printf("failed to convert discord user id to int: %s", err.Error())
					return
				}

				username, err := app.db.GetUsernameByUserDiscordId(userId)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "You are not registered")
					return
				}

				err = app.db.StoreBlikInDatabase(num, userId, username)
				if err != nil {
					log.Printf("something went wrong... %s", err.Error())
					return
				}

				// Send the new code to the Discord channel
				_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Your new code is: %d", num))
				fmt.Println("here")
				if err != nil {
					log.Printf("failed to send message to Discord channel: %s", err.Error())
				}

				break
			} else {
				log.Printf("failed to verify that blik code does not exist: %s", err.Error())
				return
			}
		}
	}
}

// Register registers users
func (app *Application) Register(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	arg := strings.Split(m.Content, " ")

	if arg[0] == "/register" {
		args := strings.Split(strings.TrimPrefix(m.Content, "/register"), " ")

		if len(args) != 3 {
			s.ChannelMessageSend(m.ChannelID, "Invalid number of arguments. Usage: /register <username> <password>")
			return
		}

		username := args[1]
		password := args[2]

		err := app.db.StoreUserInfo(m.Author.ID, username, password)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Something went wrong "+err.Error()+" ... Try again later.")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "You have been successfully registered!")
	}
}

// Stop removes the user from the database
func (app *Application) Stop(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if m.Content == "/stop" {
		args := strings.Split(strings.TrimPrefix(m.Content, "/stop"), " ")

		pass, err := bcrypt.GenerateFromPassword([]byte(args[2]), bcrypt.DefaultCost)

		err = app.db.DeleteUserInfo(args[1], string(pass))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Something went wrong "+err.Error()+" ... Try again later.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Your data has been successfully deleted!")
	}
}

// CheckUserActivity checks if the user is active
func (app *Application) CheckUserActivity(s *discordgo.Session, m *discordgo.MessageCreate) {
	userIds, err := app.db.CheckUserActivity()
	if err != nil {
		log.Printf("failed to check user activity: %s", err.Error())
		return
	}

	for _, userId := range userIds {
		userChannel, err := s.UserChannelCreate(userId)
		if err != nil {
			log.Printf("failed to create user channel: %s", err.Error())
			return
		}

		_, _ = s.ChannelMessageSend(userChannel.ID, "You have been inactive for 15 hours. Please log in to the server to avoid being kicked.")
	}
}
