package bot

import (
	"bwmarrin/discordgo"
	"fmt"
)

func handleErr(err error) {
	if err != nil {
		fmt.Println("\033[31m[FAIL]\033[0m Error: %v", err)
	}
}

func Run(token string) {
	session, err := discordgo.New("Bot " + token)
	handleErr(err)

	session.AddHandler(messageCreate)
	err = session.Open()
	defer session.Close()
	handleErr(err)

	fmt.Println("\033[32m[SUCCESS]\033[0m Bot is running")
}

func tree(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == discord.State.User.ID {
		return
	}
	  
	switch {
		case strings.Contains(message.Content, "!help"):
			help(message.ChannelID, session)
		case strings.Contains(message.Content, "!ping"):
			ping(message.ChannelID, session)
	}
}