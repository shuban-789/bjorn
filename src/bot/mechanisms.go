package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
)

func init() {
	RegisterCommand(
		&discordgo.ApplicationCommand{
			Name:                     "mech",
			Description:              "Mechanic/admin commands for the bot.",
			DefaultMemberPermissions: func() *int64 { p := int64(discordgo.PermissionAdministrator); return &p }(),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "restart",
					Description: "Restart the bot.",
					ChannelTypes: interactions.GUILDS_ONLY,
				},
			},
		},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			data := i.ApplicationCommandData()
			if len(data.Options) == 0 {
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Content: "No subcommand provided.", Flags: 1 << 6}})
				return
			}
			sub := data.Options[0]
			subName := sub.Name

			// todo: note that i.mmember is nil in dms so this will error out, use GetAuthorId instead, for now I just restrict to guilds only so we can do admin checks
			isAdminUser, err := isAdmin(s, i.GuildID, i.Member.User.ID)
			if err != nil {
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Content: "Unable to check permissions.", Flags: 1 << 6}})
				return
			}
			if !isAdminUser {
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Content: "You do not have permission to run this command.", Flags: 1 << 6}})
				return
			}

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			switch subName {
			case "restart":
				restartBot(s, i.ChannelID, i)
			default:
				interactions.SendMessage(s, i, "", "Unknown mech subcommand.")
			}
		},
	)
}

func mechcmd(session *discordgo.Session, message *discordgo.MessageCreate, i *discordgo.InteractionCreate, args []string) {
	channelId := interactions.GetChannelId(message, i)
	guildId, guildRetrieved := interactions.GetGuildId(message, i)
	authorId, authorRetrieved := interactions.GetAuthorId(message, i)

	if !authorRetrieved || !guildRetrieved {
		interactions.SendMessage(session, i, channelId, "Unable to retrieve author or guild information. This command can only be used in a server.")
		return
	}

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
