package commands

import (
	"fmt"
	"freelance-discord/models"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

var (
	quizChatId = "1101477795005222935"
)

func HasPerm(session *discordgo.Session, user *discordgo.User, channelID string, perm int64) bool {
	perms, err := session.State.UserChannelPermissions(user.ID, channelID)
	if err != nil {
		_, _ = session.ChannelMessageSend(channelID, fmt.Sprintf("Failed to retrieve perms: %s", err.Error()))
		return false
	}
	return perms&perm != 0
}

func (app *Application) CreateQuiz(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.Content == "" {
		return
	}

	args := strings.Split(m.Content, " ")
	command := args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = []string{}
	}

	if command == "/quiz" {
		if args == nil {
			s.ChannelMessageSend(m.ChannelID, "Неправильное количество аргументов. Используй: !quiz <вопрос> <правильный ответ> <опыт за правильный ответ> <$ за правильный ответ>")
			return
		}

		if len(args) < 4 {
			s.ChannelMessageSend(m.ChannelID, "Неправильное количество аргументов. Используй: !quiz <вопрос> <правильный ответ> <опыт за правильный ответ> <$ за правильный ответ>")
			return
		}

		if HasPerm(s, m.Author, m.ChannelID, discordgo.PermissionBanMembers) {
			_, _ = s.ChannelMessageSend(m.ChannelID, "You don't have permission to use this command.")
			return
		}

		question := args[0]
		answer := args[1]
		exp := args[2]
		money := args[3]

		quiz := &models.Quizes{
			AuthorId:                 m.Author.ID,
			QuizTitle:                question,
			QuizQuestion:             question,
			QuizAnswers:              answer,
			QuizExperienceReward:     exp,
			QuizServerCurrencyReward: money,
			CreatedAt:                time.Now(),
		}

		err := app.Db.CreateQuiz(quiz)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Не удалось создать викторину.")
			return
		}

		quizChat, err := s.Channel(quizChatId)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Failed to retrieve quiz chat.")
			return
		}

		s.ChannelMessageSend(quizChat.ID, `
		Викторина началась!. Впиши '!quiz' <ответ> чтобы ответить на вопрос.
		Найдите ответ на этот вопрос: 
			%s`)

	}
}
