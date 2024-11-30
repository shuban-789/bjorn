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
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages
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

	fmt.Printf("\033[33m[INFO]\033[0m Message Details: Content='%s', Author='%s', Channel='%s'\n", 
		message.Content, message.Author.Username, message.ChannelID)

	if strings.TrimSpace(message.Content) == "" {
		fmt.Println("\033[33m[INFO]\033[0m Message content is empty. Ignoring.")
		return
	}

	if strings.HasPrefix(message.Content, ">>") {
		command := strings.TrimPrefix(message.Content, ">>")
		command = strings.TrimSpace(command)

		fmt.Printf("\033[33m[INFO]\033[0m Processing command: '%s'\n", command)

		switch command {
		case "help":
			helpcmd(message.ChannelID, session)
		case "ping":
			pingcmd(message.ChannelID, session)
		default:
			session.ChannelMessageSend(message.ChannelID, "Unknown command. Use `>>help` for a list of commands.")
		}
	}
}