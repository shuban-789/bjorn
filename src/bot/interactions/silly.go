package interactions

import (
	"fmt"
	"math/rand/v2"

	"github.com/bwmarrin/discordgo"
)

var evil bool = true

var evilJokes = []string{
	"I sort fries by length before eating them.",
	"I put a set of Legos on the floor of every hallway.",
	"I return shopping carts to the wrong side of the lot.",
	"I set all your alarms six or seven minutes apart.",
	"I leave empty water bottles in the fridge.",
	"I change the language on your devices to something you don't understand.",
	"I send six seven in all the gcs for no reason.",
	"I leave the lights on in every room.",
}

func getEvilSuffix() string {
	if len(evilJokes) == 0 {
		return "I am evil today."
	}

	return fmt.Sprintf("I am being devious today. This is my latest evil action: %s", evilJokes[rand.IntN(len(evilJokes))])
}

func applyEvilToMessage(message string) string {
	if !evil {
		return message
	}

	return fmt.Sprintf("%s\n\n%s", message, getEvilSuffix())
}

func applyEvilToEmbed(embed *discordgo.MessageEmbed) {
	if !evil || embed == nil {
		return
	}

	footerText := getEvilSuffix()
	if embed.Footer != nil {
		if embed.Footer.Text == "" {
			embed.Footer.Text = footerText
			return
		}

		embed.Footer.Text = fmt.Sprintf("%s | %s", embed.Footer.Text, footerText)
		return
	}

	embed.Footer = &discordgo.MessageEmbedFooter{Text: footerText}
}

func applyEvilToEmbeds(embeds *[]*discordgo.MessageEmbed) {
	if !evil || embeds == nil {
		return
	}

	for _, embed := range *embeds {
		applyEvilToEmbed(embed)
	}
}
