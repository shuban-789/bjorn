package pagination

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const jumpModalInputId string = "page_input"

// A paginator creates paginated messages with buttons to navigate pages
// It can read and write pagination state to customids
type Paginator struct {
	// e.g., "team;awards", keep this less than 30 chars max since total customid length is 100
	CustomIDPrefix string

	// for initial page render
	Create CreatePage

	// has a function to update a page given a page number
	Update UpdatePage

	// These keys are the keys used to access extra data in PaginationState.ExtraData
	ExtraDataKeys []string
}

// CreatePage is a function that takes in the pagination state and returns the embed for that page
type CreatePage func(state PaginationState, createParams ...any) (*discordgo.MessageEmbed, error)

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
		panic("invalid pagination interaction type!")
	}
}

func (p *Paginator) GetPaginationData(state PaginationState) string {
	return  fmt.Sprintf("%d_%d", state.CurrentPage, state.TotalPages)
}

func (p *Paginator) GetExtraDataString(state PaginationState) string {
    if len(p.ExtraDataKeys) == 0 {
        return ""
    }
    values := make([]string, 0, len(p.ExtraDataKeys))
    for _, key := range p.ExtraDataKeys {
        if val, ok := state.ExtraData[key]; ok {
            values = append(values, val)
        }
    }
    return strings.Join(values, "_")
}

func (p *Paginator) GetComponentIdWithData(state PaginationState, interactionType PaginationInteractionType) string {
    retval := p.GetComponentId(interactionType) + " " + p.GetPaginationData(state)
    extra := p.GetExtraDataString(state)
    if extra != "" {
        return retval + " " + extra
    }
    return retval
}

func (p *Paginator) GetAllComponentIds() (id_prev, id_jump_button, id_next, id_jump_modal string) {
	id_prev = p.GetComponentId(PREV_BUTTON)
	id_jump_button = p.GetComponentId(JUMP_BUTTON)
	id_next = p.GetComponentId(NEXT_BUTTON)
	id_jump_modal = p.GetComponentId(JUMP_MODAL)
	return
}

func (p *Paginator) GetAllComponentIdsWithData(state PaginationState) (id_prev, id_jump, id_next, id_jump_modal string) {
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

func ParseCustomId(data []string) (currentPage int, totalPages int, extraData []string, err error) {
	paginationParts := strings.SplitN(data[0], "_", 2)
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

	if len(data) == 2 {
		extraData = strings.Split(data[1], "_")
	} else {
		extraData = []string{}
	}

	return
}

func (p *Paginator) GetStateFromCustomId(data []string) (state PaginationState, err error) {
	currentPage, totalPages, extraData, err := ParseCustomId(data)
	if err != nil {
		return
	}

	state = PaginationState{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		ExtraData:   make(map[string]string),
	}

	for i, key := range p.ExtraDataKeys {
		if i < len(extraData) { // these should be the same length but just in case
			state.ExtraData[key] = extraData[i]
		}
	}
	return
}

func (p *Paginator) CreatePaginationButtons(state PaginationState) discordgo.ActionsRow {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			&discordgo.Button{
				Emoji:    &discordgo.ComponentEmoji{Name: "⬅️"},
				Style:    discordgo.SecondaryButton,
				CustomID: p.GetComponentIdWithData(state, PREV_BUTTON),
				Disabled: state.CurrentPage == 0,
			},
			&discordgo.Button{
				Label:    "Go to Page",
				Style:    discordgo.SecondaryButton,
				CustomID: p.GetComponentIdWithData(state, JUMP_BUTTON),
				Disabled: state.TotalPages < 2,
			},
			&discordgo.Button{
				Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				Style:    discordgo.SecondaryButton,
				CustomID: p.GetComponentIdWithData(state, NEXT_BUTTON),
				Disabled: state.CurrentPage >= state.TotalPages-1,
			},
		},
	}
}
