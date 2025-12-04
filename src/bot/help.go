package bot

import (
	"github.com/bwmarrin/discordgo"
)

func helpcmd(ChannelID string, session *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "Help",
		Description: "List of commands",
		Color:       0x72cfdd,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:  "`>>help`",
				Value: "Display this message\n",
			},
			&discordgo.MessageEmbedField{
				Name:  "`>>lead [year] [event_code]`",
				Value: "Display the leaderboard for a certain event\n",
			},
			&discordgo.MessageEmbedField{
				Name:  "`>>match info [year] [event_code] [match_number]`",
				Value: "Lookup information about a certain match\n",
			},
			&discordgo.MessageEmbedField{
				Name:  "`>>match eventstart [year] [event_code]`",
				Value: "Start an active match tracker for a current even\n",
			},
			&discordgo.MessageEmbedField{
				Name:  "`>>ping`",
				Value: "Get bot response latency",
			},
			&discordgo.MessageEmbedField{
				Name:  "`>>roleme [team_id]`",
				Value: "Assign yourself a role based on your team\nnumber (San Diego FTC teams only)\n",
			},
			&discordgo.MessageEmbedField{
				Name:  "`>>team [team_id] [optional: stats, awards]`",
				Value: "Return information about a team\n",
			},
		},
	}
	sendEmbed(session, i, ChannelID, embed)
}
