package commands

import "github.com/bwmarrin/discordgo"

// CheckUserActivity checks if the user is active
func (app *Application) CheckUserActivity(s *discordgo.Session, m *discordgo.MessageCreate) {
	userIds, err := app.Db.CheckUserActivity()
	if err != nil {
		logger.Printf("failed to check user activity: %s", err.Error())
		return
	}

	for _, userId := range userIds {
		userChannel, err := s.UserChannelCreate(userId)
		if err != nil {
			logger.Printf("failed to create user channel: %s", err.Error())
			return
		}

		_, _ = s.ChannelMessageSend(userChannel.ID, "You have been inactive for 24 hours. Log in to server and enjoy your stay!")
	}
}
