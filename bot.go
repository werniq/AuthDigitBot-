package main

import (
	"fmt"
	"freelance-discord/commands"
	"freelance-discord/driver"
	"freelance-discord/models"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	//rudderfish.aternos.host:34500
	"syscall"
)

var logger = log.New(os.Stdout, "ERRORc: ", log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Printf("Error loading .env file")
	}

	bott, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		logger.Println(err)
	}

	db, err := driver.OpenDb()
	if err != nil {
		logger.Println(err)
	}

	var app *commands.Application
	app = &commands.Application{
		Db: &models.DatabaseModel{DB: db},
	}

	_, err = app.Db.Config()
	if err != nil {
		logger.Println(err)
	}

	bott.Identify.Intents = discordgo.IntentsGuildMessages
	bott.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentsGuildPresences

	//bott.AddHandler(app.GenerateBlikCode)
	bott.AddHandler(app.SaveMessage)
	bott.AddHandler(app.AntiSpamFunction)
	bott.AddHandler(app.RewardForMessages)
	bott.AddHandler(app.CheckUserIPAddressChanging)
	bott.AddHandler(app.CheckUserActivity)
	bott.AddHandler(app.OnVoiceStateUpdate)

	bott.AddHandler(app.Register)
	bott.AddHandler(app.Stop)
	bott.AddHandler(app.FindCorrectAnswersForQuiz)
	bott.AddHandler(app.CreateQuiz)

	err = bott.Open()
	if err != nil {
		fmt.Println("Error opening Discord Session, ", err)
	}
	fmt.Println("Bot is currently running. CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
