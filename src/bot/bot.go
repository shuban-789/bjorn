package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

/**
 * Returns true if there was an error, returns false otherwise.
 */
func HandleErr(err error) bool {
	if err != nil {
		fmt.Println(fail("Error: %v", err))
		return true
	}

	return false
}

var inScopeToken string

func Deploy(token string) {
	inScopeToken = token
	session, err := discordgo.New("Bot " + token)
	HandleErr(err)

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsGuildMembers

	session.AddHandler(Tree)
	session.AddHandler(slashCommandListener)
	session.AddHandler(memberJoinListener)

	err = session.Open()
	defer session.Close()
	HandleErr(err)
	fmt.Println(success("Bot is running"))

	_, err = session.ApplicationCommandBulkOverwrite(session.State.Application.ID, "", commands)
	if err != nil {
		fmt.Println(fail("Cannot register commands: %v", err))
	}
	fmt.Println(success("Application commands registered"))

	startEventUpdater(session, 2*time.Second)

	allCommands, err := session.ApplicationCommands(session.State.User.ID, "")
	HandleErr(err)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println(info("Shutting down bot..."))

	for _, cmd := range allCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, "", cmd.ID)
		if err != nil {
			log.Fatalf("Cannot delete %q command: %v", cmd.Name, err)
		}
	}
}

func slashCommandListener(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
		h(s, i)
	}
}

func Tree(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	fmt.Println(info("Message Details: Content='%s', Author='%s', Channel='%s'",
		message.Content, message.Author.Username, message.ChannelID))

	if strings.TrimSpace(message.Content) == "" {
		fmt.Println(info("Message content is empty. Ignoring."))
		return
	}

	content := strings.TrimSpace(message.Content)
	if after, ok := strings.CutPrefix(content, cmd_prefix); ok {
		content = after
		args := strings.Fields(content)
		if len(args) == 0 {
			return
		}

		cmd := strings.ToLower(args[0])
		fmt.Println(info("Processing command: '%s'", cmd))

		switch cmd {
		case "help":
			helpcmd(message.ChannelID, session, nil)
		case "ping":
			pingcmd(message.ChannelID, session, nil)
		case "team":
			teamcmd(message.ChannelID, args[1:], session, nil)
		case "roleme":
			rolemeCmd(message.ChannelID, args[1:], session, message.GuildID, message.Author.ID)
		case "match":
			matchcmd(message.ChannelID, args[1:], session, message.GuildID, message.Author.ID)
		case "lead":
			leadcmd(message.ChannelID, args[1:], session)
		case "mech":
			mechcmd(message.ChannelID, args[1:], session, message.GuildID, message.Author.ID)
		default:
			session.ChannelMessageSend(message.ChannelID, "Unknown command. Use `>>help` for a list of commands.")
		}
	}
}
