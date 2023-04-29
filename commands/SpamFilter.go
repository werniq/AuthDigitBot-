package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

// AntiSpamFunction checks if the user is spamming
func (app *Application) AntiSpamFunction(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg, err := app.Db.UserLastMessageTimestamp(m.Author.ID)
	if err != nil {
		logger.Printf("failed to get user last message timestamp: %s", err.Error())
		return
	}

	fmt.Println(time.Now().Unix())
	fmt.Println(msg.CreatedAt.Unix())
	if time.Now().Unix()-msg.CreatedAt.Unix() <= 3 {
		s.ChannelMessageSend(m.ChannelID, "Слишко много сообщений за короткий промежуток времени. Пожалуйста, не спамьте.")
	}
}
