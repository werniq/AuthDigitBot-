package commands

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

// Stop removes the user from the database
func (app *Application) Stop(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	args := strings.Split(m.Content, " ")
	command := args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = []string{}
	}

	if command == "/stop" {
		user, err := app.Db.GetUserDataFromDatabaseByUserId(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Something went wrong "+err.Error()+" ... Try again later.")
			return
		}

		if app.Db.ComparePasswords(user.Password, args[0]) != true {
			s.ChannelMessageSend(m.ChannelID, "Не верный пароль")
			return
		}

		err = app.Db.DeleteUserInfo(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Something went wrong "+err.Error()+" ... Try again later.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Ваши данные удалены из базы данных. Возвращайтесь, мы всегда Вас ждем! :)")
	}
}
