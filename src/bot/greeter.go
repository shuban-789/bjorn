package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func memberJoinListener(session *discordgo.Session, event *discordgo.GuildMemberAdd) {
	fmt.Println("\033[33m[INFO]\033[0m New member joined:", event.User.Username)
	channel, err := session.UserChannelCreate(event.User.ID)
	if err != nil {
		fmt.Println("\033[31m[FAIL]\033[0m Failed to create DM channel:", err)
		return
	}

	greet(channel.ID, session)
}

func greet(ChannelID string, session *discordgo.Session) {
	embed := &discordgo.MessageEmbed{
		Title:       "Welcome to the San Diego FTC Discord Server!",
		Description: "Get started with the information below",
		Color:       0x72cfdd,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:  "1️⃣ Get your team's role!",
				Value: "Use `>>roleme [team_id]` to get your team's role. If you are an SD team, we have your team's role :)\n",
			},
			&discordgo.MessageEmbedField{
				Name:  "2️⃣ Remember to practice Gracious Professionalism",
				Value: "The culture in this server is FIRST culture!\n",
			},
			&discordgo.MessageEmbedField{
				Name:  "3️⃣ Have fun!",
				Value: "Reach to the moderators for any help, use `>>help` to see what I can help you with.\n",
			},
		},
	}
	session.ChannelMessageSendEmbed(ChannelID, embed)
}
