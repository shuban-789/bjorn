package bot

import (
	"github.com/bwmarrin/discordgo"
)

func mechcmd(session *discordgo.Session, message *discordgo.MessageCreate, i *discordgo.InteractionCreate, args []string) {
	channelId := getChannelId(message, i)
	guildId := getGuildId(message, i)
	authorId := getAuthorId(message, i)

	if len(args) > 1 {
		sendMessage(session, i, channelId, "Please provide a subcommand (e.g., 'restart').")
		return
	}

	subCommand := args[0]

	switch subCommand {
	case "restart":
		if len(args) > 1 {
			sendMessage(session, i, channelId, "Usage: `>>mech restart`")
			return
		}

		hasPerms, err := isAdmin(session, guildId, authorId)
		if err != nil {
			sendMessage(session, i, channelId, "Unable to check permissions of user.")
			return
		}

		if hasPerms {
			sendMessage(session, i, channelId, "You do not have permission to run this command.")
			return
		}

		restartBot(session, channelId, i)
	default:
		sendMessage(session, i, channelId, "Unknown subcommand. Available subcommands: `restart`")
	}
}

func restartBot(session *discordgo.Session, channelID string, i *discordgo.InteractionCreate) {
	sendMessage(session, i, channelID, "Restarting bot...")
	session.Close()
	Deploy(inScopeToken)
}
