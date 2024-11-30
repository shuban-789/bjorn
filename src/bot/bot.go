package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
)

func HandleErr(err error) {
	if err != nil {
		fmt.Println("\033[31m[FAIL]\033[0m Error: %v", err)
	}
}

func Deploy(token string) {
	session, err := discordgo.New("Bot " + token)
	HandleErr(err)
	session.AddHandler(Tree)
	err = session.Open()
	defer session.Close()
	HandleErr(err)
	fmt.Println("\033[32m[SUCCESS]\033[0m Bot is running")

	select {}
}

func SlashCommands(session *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Shows the help menu",
		},
		{
			Name:        "ping",
			Description: "Checks the bot's response time",
		},
	}

	for _, command := range commands {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, "", command)
		if err != nil {
			fmt.Printf("Failed to create command '%s': %v\n", command.Name, err)
		}
	}
}

func Tree(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}
	
	switch {
		case strings.Contains(message.Content, ">>help"):
			helpcmd(message.ChannelID, session)
		case strings.Contains(message.Content, ">>ping"):
			pingcmd(message.ChannelID, session)
	}
}