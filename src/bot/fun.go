// A set of fun commands that don't fit well into other categories. Mostly for usage by admins and mods to mess around with people.
package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
	"github.com/shuban-789/bjorn/src/bot/util"
)

func init() {
	interactions.RegisterCommand(
		&discordgo.ApplicationCommand{
			Name:        "say",
			Description: "Have Bjorn say something in a specific channel.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "The text that Bjorn should say.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel where Bjorn should say the text.",
					Required:    true,
					ChannelTypes: []discordgo.ChannelType{
						discordgo.ChannelTypeGuildText,
						discordgo.ChannelTypeGuildPublicThread,
						discordgo.ChannelTypeGuildPrivateThread,
					},
				},
			},
		},
		sayCommandHandler,
	)
}

func sayCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fmt.Println(util.Info("Received /say command from %s#%s", i.Member.User.Username, i.Member.User.Discriminator))

	data := i.ApplicationCommandData()
	text := data.Options[0].StringValue()
	channel := data.Options[1].ChannelValue(s)
	fmt.Println(util.Info("Sending message to channel %s: %s", channel.ID, text))

	_, err := s.ChannelMessageSend(channel.ID, text)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Failed to send message: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Message sent successfully.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}
