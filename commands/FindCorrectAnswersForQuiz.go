package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

func (app *Application) FindCorrectAnswersForQuiz(s *discordgo.Session, m *discordgo.MessageCreate) {
	var err error
	var title string
	if m.Author.Bot || m.GuildID == "" || m.Content == "" {
		return
	}

	args := strings.Split(m.Content, " ")
	command := args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = []string{}
	}

	if command == "!answer" {
		if m.ChannelID != "1101477795005222935" {
			s.ChannelMessageSend(m.ChannelID, "Эта команда доступна только в канале #викторины")
		}

		title, err = app.Db.CheckIfMessageIsAnswer(args[0], m.Author.ID)
		if title != "" {
			_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Правильный ответ к опросу %s! Поздравляю %v! Валюта и опыт зачислены на ваш счет!", title, m.Author.Mention()))
			if err != nil {
				log.Printf("failed to send message: %s", err.Error())
				return
			}
		}
	}
}
