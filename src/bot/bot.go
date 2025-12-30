package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
	"github.com/shuban-789/bjorn/src/bot/util"
)

// This should be blank in production. However, it takes ~2 hours for commands
// to propagate globally which is very cooked, so you should set this to your
//
//	test server's ID for testing so that commands register instantly.
var GuildId string = os.Getenv("GUILD_ID")

/**
 * Returns true if there was an error, returns false otherwise.
 */
func HandleErr(err error) bool {
	if err != nil {
		fmt.Println(util.Fail("Error: %v", err))
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
	fmt.Println(util.Success("Bot is running"))

	_, err = session.ApplicationCommandBulkOverwrite(session.State.Application.ID, GuildId, interactions.Commands)
	if err != nil {
		fmt.Println(util.Fail("Cannot register commands: %v", err))
	}
	fmt.Println(util.Success("Application commands registered"))

	startMatchEventUpdater(session, 2*time.Second)

	allCommands, err := session.ApplicationCommands(session.State.User.ID, GuildId)
	HandleErr(err)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println(util.Info("Shutting down bot..."))

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
		if h, ok := interactions.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	
	case discordgo.InteractionApplicationCommandAutocomplete:
		// first see if a custom one exists
		if h, ok := interactions.AutocompleteHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
			return
		}

		interactions.HandleAutocomplete(s, i)

	case discordgo.InteractionMessageComponent:
		authorid, _ := interactions.GetAuthorName(nil, i)
		fmt.Println(util.Info("Received component interaction: CustomID='%s', User='%s'", i.MessageComponentData().CustomID, authorid))
		// NOTE: See src/bot/README.md for the format used in custom IDs
		fields := strings.Fields(i.MessageComponentData().CustomID)
		if h, ok := interactions.ComponentHandlers[fields[0]]; ok {
			if len(fields) == 1 {
				h(s, i, []string{""})
			} else {
				h(s, i, fields[1:])
			}
		}

	case discordgo.InteractionModalSubmit:
		authorid, _ := interactions.GetAuthorName(nil, i)
		fmt.Println(util.Info("Received modal submit interaction: CustomID='%s', User='%s'", i.ModalSubmitData().CustomID, authorid))

		// NOTE: See src/bot/README.md for the format used in custom IDs
		var modalData discordgo.ModalSubmitInteractionData = i.ModalSubmitData()
		fields := strings.Fields(modalData.CustomID)
		if h, ok := interactions.ModalHandlers[fields[0]]; ok {
			if len(fields) == 1 { // no extra data
				h(s, i, []string{}, modalData)
			} else {
				h(s, i, fields[1:], modalData)
			}
		}
	}
}

func Tree(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	fmt.Println(util.Info("Message Details: Content='%s', Author='%s', Channel='%s'",
		message.Content, message.Author.Username, message.ChannelID))

	if strings.TrimSpace(message.Content) == "" {
		fmt.Println(util.Info("Message content is empty. Ignoring."))
		return
	}

	content := strings.TrimSpace(message.Content)
	if after, ok := strings.CutPrefix(content, interactions.CmdPrefix); ok {
		content = after
		args := strings.Fields(content)
		if len(args) == 0 {
			return
		}

		cmd := strings.ToLower(args[0])
		fmt.Println(util.Info("Processing command: '%s' with arguments %s", cmd, args[1:]))

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
