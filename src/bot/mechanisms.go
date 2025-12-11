package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
)

func mechcmd(session *discordgo.Session, message *discordgo.MessageCreate, i *discordgo.InteractionCreate, args []string) {
	channelId := interactions.GetChannelId(message, i)
	guildId := interactions.GetGuildId(message, i)
	authorId := interactions.GetAuthorId(message, i)

	if len(args) > 1 {
		interactions.SendMessage(session, i, channelId, "Please provide a subcommand (e.g., 'restart').")
		return
	}

	subCommand := args[0]

	switch subCommand {
	case "restart":
		if len(args) > 1 {
			interactions.SendMessage(session, i, channelId, "Usage: `>>mech restart`")
			return
		}

		hasPerms, err := isAdmin(session, guildId, authorId)
		if err != nil {
			interactions.SendMessage(session, i, channelId, "Unable to check permissions of user.")
			return
		}

		if hasPerms {
			interactions.SendMessage(session, i, channelId, "You do not have permission to run this command.")
			return
		}

		restartBot(session, channelId, i)
	default:
		interactions.SendMessage(session, i, channelId, "Unknown subcommand. Available subcommands: `restart`")
	}
}

func restartBot(session *discordgo.Session, channelID string, i *discordgo.InteractionCreate) {
	interactions.SendMessage(session, i, channelID, "Restarting bot...")
	session.Close()
	Deploy(inScopeToken)
}
