package bot

import "github.com/bwmarrin/discordgo"

func helpcmd(ChannelID string, session *discordgo.Session) {
	embed := &discordgo.MessageEmbed{
		Title:       "Help",
		Description: "List of commands",
		Color:       0x72cfdd,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:  ">>help",
				Value: "Display this message",
			},
			&discordgo.MessageEmbedField{
				Name:  ">>ping",
				Value: "Get bot response latency",
			},
			&discordgo.MessageEmbedField{
				Name:  ">>team [team_id] [optional: stats, awards]",
				Value: "Return information about a team",
			},
			&discordgo.MessageEmbedField{
				Name:  ">>roleme [team_id]",
				Value: "Assign yourself a role based on your team\nnumber (San Diego FTC teams only)",
			},
		},
	}
	session.ChannelMessageSendEmbed(ChannelID, embed)
}
