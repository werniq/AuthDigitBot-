package commands

// GenerateBlikCode generates blik code
func (app *Application) GenerateBlikCode(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if m.Content == "/blik" {
		// Set up a timer that runs the code generation logic every 30 seconds
		timer := time.NewTicker(30 * time.Second)
		defer timer.Stop()

		for {
	   	// Generate a new code
			num := rand.Intn(10000000) + 1

			// Check if the code already exists in the database
			err := app.Db.VerifyThatBlikNotExists(num)

			if err == nil {
				// Store the code in the database
				userId, err := strconv.Atoi(m.Author.ID)
				if err != nil {
					logger.Printf("failed to convert discord user id to int: %s", err.Error())
					return
				}

				username, err := app.Db.GetUsernameByUserDiscordId(userId)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "You are not registered")
					return
				}

				err = app.Db.StoreBlikInDatabase(num, userId, username)
				if err != nil {
					logger.Printf("something went wrong... %s", err.Error())
					return
				}

				// Send the new code to the Discord channel
				_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Your new code is: %d", num))
				fmt.Println("here")
				if err != nil {
					logger.Printf("failed to send message to Discord channel: %s", err.Error())
				}

				break
			} else {
				logger.Printf("failed to verify that blik code does not exist: %s", err.Error())
				return
			}
		}
  }
}
