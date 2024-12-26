package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

/**
 * Returns true if there was an error, returns false otherwise.
 */
func HandleErr(err error) bool {
	if err != nil {
		fmt.Println("\033[31m[FAIL]\033[0m Error: %v", err)
		return true
	}

	return false
}

func Deploy(token string) {
	session, err := discordgo.New("Bot " + token)
	HandleErr(err)
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages
	session.AddHandler(Tree)
	session.AddHandler(memberJoinListener)
	err = session.Open()
	defer session.Close()
	HandleErr(err)
	fmt.Println("\033[32m[SUCCESS]\033[0m Bot is running")

	startEventUpdater(session, 2*time.Minute)

	select {}
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
		args := strings.Fields(message.Content)
		if len(args) == 0 {
			return
		}
		fmt.Printf("\033[33m[INFO]\033[0m Processing command: '%s'\n", args[0])

		switch args[0] {
		case ">>help":
			helpcmd(message.ChannelID, session)
		case ">>ping":
			pingcmd(message.ChannelID, session)
		case ">>team":
			teamcmd(message.ChannelID, args[1:], session)
		case ">>roleme":
			rolemeCmd(message.ChannelID, args[1:], session, message.GuildID, message.Author.ID)
		case ">>match":
			matchcmd(message.ChannelID, args[1:], session, message.GuildID, message.Author.ID)
		case ">>lead":
			leadcmd(message.ChannelID, args[1:], session)
		default:
			session.ChannelMessageSend(message.ChannelID, "Unknown command. Use `>>help` for a list of commands.")
		}
	}
}
