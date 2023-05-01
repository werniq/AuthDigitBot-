package commands

import (
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

	if time.Now().Unix()-msg.CreatedAt.Unix() <= 2 {
		s.ChannelMessageSend(m.ChannelID, "Слишко много сообщений за короткий промежуток времени. Пожалуйста, не спамьте.")
	}
}
