package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

var (
	avaliableLevel = []string{"1 уровень", "2 уровень", "3 уровень", "4 уровень", "5 уровень"}
)

// Register registers users
func (app *Application) Register(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if m.Type == discordgo.MessageTypeGuildMemberJoin {
		channel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			logger.Printf("failed to create user channel: %s", err.Error())
			return
		}

		welcomeMessage := fmt.Sprintf("Приветствуем на сервере, <%v>! Для того, чтобы получить доступ к каналам, вам необходимо зарегистрироваться. Для этого введите команду /register <username> <password>", m.Author.Mention())

		s.ChannelMessageSend(channel.ID, welcomeMessage)
	}
}

func (app *Application) ListenDmRegisterCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	arg := strings.Split(m.Content, " ")

	if arg[0] == "/register" {
		if m.GuildID != "" {
			s.ChannelMessageSend(m.ChannelID, "Можно регистрироваться только в личных сообщениях.")
			return
		}
		args := strings.Split(strings.TrimPrefix(m.Content, "/register"), " ")

		if len(args) != 3 {
			s.ChannelMessageSend(m.ChannelID, "Неправильное количество аргументов. Пожалуйста, введите команду в формате /register <username> <password>")
			return
		}

		username := args[1]
		password := args[2]

		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Вы не состоите в гильдии. Ссылка на гильдию: https://discord.gg/R7JMvfqZ")
			return
		}

		if member == nil {
			s.ChannelMessageSend(m.ChannelID, "Вы не состоите в гильдии. Ссылка на гильдию: https://discord.gg/R7JMvfqZ")
			return
		}

		level := member.Roles[0]
		role := member.Roles[1]

		if level == "" || role == "" {
			s.ChannelMessageSend(m.ChannelID, "Попросите администратора сначала выдать Вам роль и уровень :)")
			return
		}

		ok := false
		swappedOk := false

		for i := 0; i < len(avaliableLevel); i++ {
			if level == avaliableLevel[i] {
				ok = true
				break
			} else if role == avaliableLevel[i] {
				swappedOk = true
				break
			}
		}

		if swappedOk {
			level = role
			role = member.Roles[0]
		} else if !ok {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Unknown level. Please contact the server administrator to assign you one of available roles.")
			return
		}

		if !ok {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Unknown role. Please contact the server administrator.")
			return
		}

		err = app.Db.StoreUserInfo(m.Author.ID, username, password, role, level)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Something went wrong "+err.Error()+" ... Try again later.")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "You have been successfully registered!")
	}
}
