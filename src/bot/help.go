package bot

import "github.com/bwmarrin/discordgo"

func helpcmd(ChannelID string, session *discordgo.Session) {
	embed := &discordgo.MessageEmbed{
		Title: "Help",
		Description: "List of commands",
		Color: 0x72cfdd,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name: ">>help",
				Value: "Display this message",
			},
			&discordgo.MessageEmbedField{
				Name: ">>ping",
				Value: "Bot response latency",
			},
		},
	}
	session.ChannelMessageSendEmbed(ChannelID, embed)
}