package bot

import (
	"github.com/bwmarrin/discordgo"
)

func mechcmd(channelID string, args []string, session *discordgo.Session, guildId string, authorID string) {
	if len(args) > 1 {
		session.ChannelMessageSend(channelID, "Please provide a subcommand (e.g., 'restart').")
		return
	}

	subCommand := args[0]

	switch subCommand {
	case "restart":
		if len(args) > 1 {
			session.ChannelMessageSend(channelID, "Usage: `>>mech restart`")
			return
		}

		hasPerms, err := isAdmin(session, guildId, authorID)
		if err != nil {
			session.ChannelMessageSend(channelID, "Unable to check permissions of user.")
			return
		}

		if hasPerms {
			session.ChannelMessageSend(channelID, "You do not have permission to run this command.")
			return
		}

		restartBot(session, channelID)
	default:
		session.ChannelMessageSend(channelID, "Unknown subcommand. Available subcommands: `restart`")
	}
}

func restartBot(session *discordgo.Session, channelID string) {
	session.ChannelMessageSend(channelID, "Restarting bot...")
	session.Close()
	Deploy(inScopeToken)
}
