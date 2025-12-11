package interactions

import (
	"github.com/bwmarrin/discordgo"
)

func CreatePaginationButtons(totalPages int, currentPage int, id_prev, id_next string) discordgo.ActionsRow {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			&discordgo.Button{
				Label:    "Previous",
				Style:    discordgo.PrimaryButton,
				CustomID: id_prev,
				Disabled: currentPage == 0,
			},
			&discordgo.Button{
				Label:    "Next",
				Style:    discordgo.PrimaryButton,
				CustomID: id_next,
				Disabled: currentPage >= totalPages-1,
			},
		},
	}
}
