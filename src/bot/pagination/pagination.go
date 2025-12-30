package pagination

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// A paginator creates paginated messages with buttons to navigate pages
// It can read and write pagination state to customids
type Paginator[T any] struct {
	// e.g., "team;awards", keep this less than 30 chars max since total customid length is 100
	CustomIDPrefix string

	// for initial page render
	Create CreatePage

	// has a function to update a page given a page number
	Update UpdatePage

	ExtraDataHandler func([]string) (T, error)
}

// CreatePage is a function that takes in the pagination state and returns the embed for that page
type CreatePage func(state PaginationState) (*discordgo.MessageEmbed, error)

// UpdatePage is a function that takes in the pagination state and an embed to modify, and returns the modified embed
type UpdatePage func(state PaginationState, embed *discordgo.MessageEmbed) (*discordgo.MessageEmbed, error)

// PaginationState holds data about the pagination state, retrieved from customid
type PaginationState struct {
	TotalPages  int
	CurrentPage int

	// extra data to be stored in customid
	ExtraData map[string]string
}

type PaginationInteractionType int

const (
	PREV_BUTTON PaginationInteractionType = iota
	JUMP_BUTTON
	NEXT_BUTTON
	JUMP_MODAL
)

const jumpModalInputId string = "page_input"

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
func (p *Paginator) RenderCurrentPage(state PaginationState) (*discordgo.MessageEmbed, error) {
	return p.Renderer.RenderPage(state.CurrentPage)
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

func (p *Paginator) GetPaginationData(state PaginationState) string {
	return  fmt.Sprintf("%d_%d", state.CurrentPage, state.TotalPages)
}

func (p *Paginator) GetExtraDataString() string {
	return  fmt.Sprintf(" %s",  fmt.Sprint(strings.Join(p.ExtraData, "_")))
}

func (p *Paginator) GetComponentIdWithData(state PaginationState, interactionType PaginationInteractionType) string {
	return p.GetComponentId(interactionType) + " " + p.GetPaginationData(state) + " " + p.GetExtraDataString()
}

func (p *Paginator) GetAllComponentIds(state PaginationState) (id_prev, id_jump, id_next, id_jump_modal string) {
	id_prev = p.GetComponentIdWithData(state, PREV_BUTTON)
	id_jump = p.GetComponentIdWithData(state, JUMP_BUTTON)
	id_next = p.GetComponentIdWithData(state, NEXT_BUTTON)
	id_jump_modal = p.GetComponentIdWithData(state, JUMP_MODAL)
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

func (p *Paginator) CreatePaginationButtons() discordgo.ActionsRow {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			&discordgo.Button{
				Emoji:    &discordgo.ComponentEmoji{Name: "⬅️"},
				Style:    discordgo.SecondaryButton,
				CustomID: p.GetComponentIdWithData(PREV_BUTTON),
				Disabled: p.CurrentPage == 0,
			},
			&discordgo.Button{
				Label:    "Go to Page",
				Style:    discordgo.SecondaryButton,
				CustomID: p.GetComponentIdWithData(JUMP_BUTTON),
				Disabled: p.TotalPages < 2,
			},
			&discordgo.Button{
				Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				Style:    discordgo.SecondaryButton,
				CustomID: p.GetComponentIdWithData(NEXT_BUTTON),
				Disabled: p.CurrentPage >= p.TotalPages-1,
			},
		},
	}
}

func (p *Paginator) SendPaginatedMessage(s *discordgo.Session, i *discordgo.InteractionCreate, channelID string) (err error) {
	embed, err := p.RenderCurrentPage()
	if err != nil {
		return
	}
	embeds := []*discordgo.MessageEmbed{embed}

	components := []discordgo.MessageComponent{
		p.CreatePaginationButtons(),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     embeds,
			Components: components,
		},
	})
	return nil
}