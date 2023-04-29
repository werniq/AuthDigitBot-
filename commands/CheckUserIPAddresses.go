package commands

import "github.com/bwmarrin/discordgo"

// CheckUserIPAddressChanging checks if the user is active
func (app *Application) CheckUserIPAddressChanging(s *discordgo.Session, m *discordgo.MessageCreate) {
	usernames, err := app.Db.CheckUserIPAddressChanging()
	if err != nil {
		logger.Printf("failed to check user activity: %s", err.Error())
		return
	}

	userIds, err := app.Db.ConvertUsernamesToUserIds(usernames)
	if err != nil {
		logger.Printf("failed to convert usernames to user ids: %s", err.Error())
		return
	}

	for _, userId := range userIds {
		userChannel, err := s.UserChannelCreate(userId)
		if err != nil {
			logger.Printf("failed to create user channel: %s", err.Error())
			return
		}
		_, _ = s.ChannelMessageSend(userChannel.ID, "Log in attempt from a new IP address. If this was you, please ignore this message. If this wasn't you, please change your password.")
	}

}
