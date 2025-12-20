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

// This should be blank in production. However, it takes ~2 hours for commands
// to propagate globally which is very cooked, so you should set this to your
//
//	test server's ID for testing so that commands register instantly.
var GuildId string = "784133390131134504"

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
	session.AddHandler(interactionCreateHandler)
	session.AddHandler(memberJoinListener)

	err = session.Open()
	defer session.Close()
	HandleErr(err)
	fmt.Println(success("Bot is running"))

	_, err = session.ApplicationCommandBulkOverwrite(session.State.Application.ID, GuildId, commands)
	if err != nil {
		fmt.Println(fail("Cannot register commands: %v", err))
	}
	fmt.Println(success("Application commands registered"))

	startEventUpdater(session, 2*time.Second)

	allCommands, err := session.ApplicationCommands(session.State.User.ID, GuildId)
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

func interactionCreateHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	
	case discordgo.InteractionApplicationCommandAutocomplete:
		if h, ok := autocompleteHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}

	case discordgo.InteractionMessageComponent:
		fmt.Println(info("Received component interaction: CustomID='%s', User='%s'", i.MessageComponentData().CustomID, i.Member.User.Username))
		// NOTE: See src/bot/README.md for the format used in custom IDs
		fields := strings.Fields(i.MessageComponentData().CustomID)
		if h, ok := componentHandlers[fields[0]]; ok {
			if len(fields) == 1 {
				h(s, i, "")
			} else {
				h(s, i, fields[1])
			}
		}
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
		fmt.Println(info("Processing command: '%s' with arguments %s", cmd, args[1:]))

		switch cmd {
		case "help":
			helpcmd(session, message, nil)
		case "ping":
			pingcmd(session, message, nil)
		case "team":
			teamcmd(session, message, nil, args[1:])
		case "roleme":
			rolemeCmd(session, message, nil, args[1:])
		case "match":
			matchcmd(session, message, nil, args[1:])
		case "lead":
			leadcmd(session, message, nil, args[1:])
		case "mech":
			mechcmd(session, message, nil, args[1:])
		default:
			session.ChannelMessageSend(message.ChannelID, "Unknown command. Use `>>help` for a list of commands.")
		}
	}
}
