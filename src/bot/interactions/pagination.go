package interactions

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Paginator struct {
	TotalPages  int
	CurrentPage int

	// e.g., "team;awards", keep this less than 30 chars max since total customid length is 100
	CustomIDPrefix string

	// has a function to render a page given a page number
	Renderer PageRenderer

	// extra data to be stored in customid
	ExtraData []string
}

type PageRenderer interface {
	RenderPage(pageNum int) ([]*discordgo.MessageEmbed, error)
}

type PaginationInteractionType int

const (
	PREV_BUTTON PaginationInteractionType = iota
	JUMP_BUTTON
	NEXT_BUTTON
	JUMP_MODAL
)

var interactionName = map[PaginationInteractionType]string{
    PREV_BUTTON: "prev_button",
    JUMP_BUTTON: "jump_button",
    NEXT_BUTTON: "next_button",
    JUMP_MODAL:  "jump_modal",
}
func (pit PaginationInteractionType) String() string {
    return interactionName[pit]
}

// the customids will have 3 parts, the button name, the pagination data (page number, total pages), and extra data if any
// 
// e.g., "team;awards_pb 2_5 22105" for previous button on page 2 of 5 for team 22105's awards
func (p *Paginator) RenderCurrentPage() ([]*discordgo.MessageEmbed, error) {
	return p.Renderer.RenderPage(p.CurrentPage)
}

func (p *Paginator) GetComponentId(interactionType PaginationInteractionType) string {
	switch interactionType {
	case PREV_BUTTON:
		return p.CustomIDPrefix + "_pb"
	case JUMP_BUTTON:
		return p.CustomIDPrefix + "_jb"
	case NEXT_BUTTON:
		return p.CustomIDPrefix + "_nb"
	case JUMP_MODAL:
		return p.CustomIDPrefix + "_jm"
	default:
		return ""
	}
}

func (p *Paginator) GetPaginationData() string {
	return  fmt.Sprintf("%d_%d", p.CurrentPage, p.TotalPages)
}

func (p *Paginator) GetExtraDataString() string {
	return  fmt.Sprintf(" %s",  fmt.Sprint(strings.Join(p.ExtraData, "_")))
}

func (p *Paginator) GetComponentIdWithData(interactionType PaginationInteractionType) string {
	return p.GetComponentId(interactionType) + " " + p.GetPaginationData() + " " + p.GetExtraDataString()
}

func (p *Paginator) GetAllComponentIds() (id_prev, id_jump, id_next, id_jump_modal string) {
	id_prev = p.GetComponentIdWithData(PREV_BUTTON)
	id_jump = p.GetComponentIdWithData(JUMP_BUTTON)
	id_next = p.GetComponentIdWithData(NEXT_BUTTON)
	id_jump_modal = p.GetComponentIdWithData(JUMP_MODAL)
	return
}

func getComponentType(customIdStart string) (PaginationInteractionType, error) {
	if strings.HasSuffix(customIdStart, "_pb") {
		return PREV_BUTTON, nil
	} else if strings.HasSuffix(customIdStart, "_nb") {
		return NEXT_BUTTON, nil
	} else if strings.HasSuffix(customIdStart, "_jb") {
		return JUMP_BUTTON, nil
	} else if strings.HasSuffix(customIdStart, "_jm") {
		return JUMP_MODAL, nil
	} else {
		return -1, fmt.Errorf("invalid pagination action")
	}
}

func ParseCustomId(data string) (interactionType PaginationInteractionType, currentPage int, totalPages int, extraData []string, err error) {
	parts := strings.SplitN(data, " ", 3)
	if len(parts) < 2 {
		err = fmt.Errorf("invalid custom ID format")
		return
	}

	interactionType, err = getComponentType(parts[0])
	if err != nil {
		return
	}

	paginationParts := strings.SplitN(parts[1], "_", 2)
	if len(paginationParts) != 2 {
		err = fmt.Errorf("invalid pagination data format")
		return
	}
	_, err = fmt.Sscanf(paginationParts[0], "%d", &currentPage)
	if err != nil {
		return
	}

	_, err = fmt.Sscanf(paginationParts[1], "%d", &totalPages)
	if err != nil {
		return
	}

	if len(parts) == 3 {
		extraData = strings.Split(parts[2], "_")
	} else {
		extraData = []string{}
	}

	return
}

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
