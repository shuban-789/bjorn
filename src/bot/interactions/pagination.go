package interactions

import (
	"github.com/bwmarrin/discordgo"
)

func CreatePaginationButtons(totalPages int, currentPage int, id_prev, id_jump, id_next string) discordgo.ActionsRow {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			&discordgo.Button{
				Emoji:    &discordgo.ComponentEmoji{Name: "⬅️"},
				Style:    discordgo.SecondaryButton,
				CustomID: id_prev,
				Disabled: currentPage == 0,
			},
			&discordgo.Button{
				Label:    "Go to Page",
				Style:    discordgo.SecondaryButton,
				CustomID: id_jump,
				Disabled: totalPages < 2,
			},
			&discordgo.Button{
				Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				Style:    discordgo.SecondaryButton,
				CustomID: id_next,
				Disabled: currentPage >= totalPages-1,
			},
		},
	}
}
