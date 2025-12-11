// This file contains utility functions to abstract over interactions vs messages so that code for
// commands can be shared between both.

package interactions

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func GetGuildId(message *discordgo.MessageCreate, i *discordgo.InteractionCreate) string {
	if i != nil {
		return i.GuildID
	}
	if message != nil {
		return message.GuildID
	}
	panic("Both message and interaction are nil in getGuildId")
}

func GetAuthorId(message *discordgo.MessageCreate, i *discordgo.InteractionCreate) string {
	if i != nil {
		return i.Member.User.ID
	}
	if message != nil {
		return message.Author.ID
	}
	panic("Both message and interaction are nil in getAuthorId")
}

func GetChannelId(message *discordgo.MessageCreate, i *discordgo.InteractionCreate) string {
	if i != nil {
		return i.ChannelID
	}
	if message != nil {
		return message.ChannelID
	}
	panic("Both message and interaction are nil in getChannelId")
}

func SendEmbed(session *discordgo.Session, i *discordgo.InteractionCreate, channelID string, embed *discordgo.MessageEmbed) {
	if i != nil {
		_, err := session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}})

		if err != nil {
			msg := fmt.Sprintf("Failed to send embed: %v", err)
			session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		}
	} else {
		_, err := session.ChannelMessageSendEmbed(channelID, embed)

		if err != nil {
			session.ChannelMessageSend(channelID, fmt.Sprintf("Failed to send embed: %v", err))
		}
	}
}

func SendMessage(session *discordgo.Session, i *discordgo.InteractionCreate, channelID string, message string) {
	if i != nil {
		_, err := session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &message})
		// err := session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 	Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 	Data: &discordgo.InteractionResponseData{
		// 		Content: message,
		// 	},
		// })
		if err != nil {
			msg := fmt.Sprintf("Failed to send message: %v", err)
			session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		}
	} else {
		_, err := session.ChannelMessageSend(channelID, message)
		if err != nil {
			session.ChannelMessageSend(channelID, fmt.Sprintf("Failed to send message: %v", err))
		}
	}
}

// Returns whether or not the message was sent successfully
func SendMessageComplex(session *discordgo.Session, i *discordgo.InteractionCreate, channelID string, message string, components *[]discordgo.MessageComponent, embeds *[]*discordgo.MessageEmbed) bool {
	if i != nil {
		_, err := session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &message, Components: components, Embeds: embeds})
		// err := session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 	Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 	Data: &discordgo.InteractionResponseData{
		// 		Content: message,
		// 	},
		// })
		if err != nil {
			msg := fmt.Sprintf("Failed to send message: %v", err)
			session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
			return false
		}
	} else {
		_, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Content:    message,
			Components: *components,
			Embeds:     *embeds,
		})

		if err != nil {
			session.ChannelMessageSend(channelID, fmt.Sprintf("Failed to send message: %v", err))
			return false
		}
	}

	return true
}
