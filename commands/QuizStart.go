package commands

import (
	"fmt"
	"freelance-discord/models"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

var (
	quizChatId = "1064093774852141137"
)

func HasPerm(session *discordgo.Session, user *discordgo.User, guildID string, channelID string, perm int64) bool {
	_, err := session.State.Guild(guildID)
	if err != nil {
		_, _ = session.ChannelMessageSend(channelID, fmt.Sprintf("Failed to retrieve guild: %s", err.Error()))
		return false
	}
	member, err := session.State.Member(guildID, user.ID)
	if err != nil {
		_, _ = session.ChannelMessageSend(channelID, fmt.Sprintf("Failed to retrieve member: %s", err.Error()))
		return false
	}
	return member.Permissions&perm == perm
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

		if HasPerm(s, m.Author, m.GuildID, m.ChannelID, discordgo.PermissionBanMembers) {
			s.ChannelMessageSend(m.ChannelID, "You don't have permission to use this command.")
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
			s.ChannelMessageSend(m.ChannelID, "Не удалось создать викторину."+err.Error())
			return
		}

		quizChat, err := s.Channel(quizChatId)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Failed to retrieve quiz chat.")
			return
		}

		s.ChannelMessageSend(quizChat.ID, fmt.Sprintf(`
		Викторина началась! Впиши '!quiz' <ответ> чтобы ответить на вопрос.
		Найдите ответ на этот вопрос: 
			%s`, quiz.QuizQuestion+"?"))

	}
}
