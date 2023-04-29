package commands

import (
	"freelance-discord/models"
	"github.com/bwmarrin/discordgo"
)

// SaveMessage saves message to database
func (app *Application) SaveMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.GuildID == "" || m.Content == "" {
		return
	}

	if m.ID == "" || m.Author.ID == "" || m.ChannelID == "" {
		return
	}

	msg := &models.Message{
		MessageId: m.ID,
		UserId:    m.Author.ID,
		ChannelId: m.ChannelID,
		Content:   m.Content,
		CreatedAt: m.Timestamp,
	}

	err := app.Db.SaveMessageToDatabase(msg)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "error saving message into database: %s"+err.Error())
	}
}
