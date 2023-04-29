package commands

import (
	"github.com/bwmarrin/discordgo"
	"time"
)

var (
	timers = map[string]time.Time{}
)

// RewardForMessages rewards users for messages
func (app *Application) RewardForMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	users, err := app.Db.SelectUsersWithHundredMessages()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Ошибка при получении пользователей с 100 сообщениями")
		return
	}

	err = app.Db.ChargeUsersWithMoney(users)
	if err != nil {
		logger.Printf("failed to reward for messages: %s", err.Error())
		return
	}
}

func (app *Application) OnVoiceStateUpdate(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	if vs.ChannelID != "" {
		timer := time.Now()
		timers[vs.UserID] = timer
	} else {
		if timer, ok := timers[vs.UserID]; ok {
			duration := time.Since(timer)

			seconds := int(duration.Seconds())
			err := app.Db.StoreSecondsInVoiceChar(vs.UserID, seconds)
			if err != nil {
				s.ChannelMessageSend(vs.ChannelID, "Ошибка при сохранении времени в войсе")
				return
			}
		}
	}
}
