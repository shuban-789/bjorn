package bot

import "bwmarrin/discordgo"

func help(ChannelID string, session *discordgo.Session) {
	embed := &discordgo.MessageEmbed{
		Title: "Help",
		Description: "List of commands",
		Color: 0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name: "!help",
				Value: "Display this message",
			},
			&discordgo.MessageEmbedField{
				Name: "!ping",
				Value: "Ping the bot",
			},
		},
	}
	session.ChannelMessageSendEmbed(ChannelID, embed)
}