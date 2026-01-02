/*
For fun I asked an AI to generate a model of the flow and it was actually kinda neat so I'm putting it here:

User clicks button
        │
        ▼
┌─────────────────────────┐
│ interactions.handlers   │  Routes by custom ID prefix
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Paginator.pageLeftRight │  Parses state from custom ID
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Paginator.Update        │  Calls user-defined callback
│ (updateAwardsEmbed)     │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│ Paginator.editMessage   │  Sends Discord API response
└─────────────────────────┘

Do note that this is only the flow for when you press the previous/next arrows though not the modal but the idea is the same thing basiclaly

Also Paginator is a struct where Update is a field of type func, and updateAwardsEmbed is assigned
to that so that you can have diff ways of generating the embeds for diff commands
*/

package pagination

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
	"github.com/shuban-789/bjorn/src/bot/util"
)

func (p *Paginator[T]) Register() {
	id_prev, id_jump_button, id_next, id_jump_modal := p.GetAllComponentIds()

	interactions.RegisterComponentHandler(id_prev, func(s *discordgo.Session, ic *discordgo.InteractionCreate, data []string) {
		err := p.pageLeftRight(s, ic, data, -1)
		if err != nil {
			fmt.Print(util.Fail(err.Error()))
		}
	})

	interactions.RegisterComponentHandler(id_jump_button, func(s *discordgo.Session, ic *discordgo.InteractionCreate, data []string) {
		err := p.launchJumpModal(s, ic, data)
		if err != nil {
			fmt.Print(util.Fail(err.Error()))
		}
	})

	interactions.RegisterModalHandler(id_jump_modal, func(s *discordgo.Session, i *discordgo.InteractionCreate, id_data []string, modal_data discordgo.ModalSubmitInteractionData) {
		err := p.handleJumpModalSubmit(s, i, id_data, modal_data)
		if err != nil {
			fmt.Print(util.Fail(err.Error()))
		}
	})

	interactions.RegisterComponentHandler(id_next, func(s *discordgo.Session, ic *discordgo.InteractionCreate, data []string) {
		err := p.pageLeftRight(s, ic, data, 1)
		if err != nil {
			fmt.Print(util.Fail(err.Error()))
		}
	})
}

func (p *Paginator[T]) pageLeftRight(s *discordgo.Session, ic *discordgo.InteractionCreate, data []string, delta int) error {
	state, err := p.GetStateFromCustomId(data)
	if err != nil {
		return err
	}
	
	state.CurrentPage = util.Clamp(state.CurrentPage+delta, 0, state.TotalPages-1)
	return p.editMessage(s, ic, state)
}


func (p *Paginator[T]) handleJumpModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate, id_data []string, modal_data discordgo.ModalSubmitInteractionData) error {
	state, err := p.GetStateFromCustomId(id_data)
	if err != nil {
		return err
	}

	input, ok := interactions.GetComponentWithId(modal_data.Components, jumpModalInputId).(*discordgo.TextInput)
	if !ok {
		return fmt.Errorf("failed to get page input component")
	}

	pageInput := input.Value
	pageNum, err := strconv.Atoi(pageInput)
	if err != nil || pageNum < 1 || pageNum > state.TotalPages {
		err := interactions.SendEphemeralMessage(s, i, fmt.Sprintf("Invalid page number. Please enter a number between 1 and %d (including the end values).", state.TotalPages))
		if err != nil {
			return fmt.Errorf("failed to send ephemeral error message %v", err)
		}
		return nil
	}
	pageNum -= 1 // convert to 0-indexed paging

	state.CurrentPage = pageNum
	return p.editMessage(s, i, state)
}

func (p *Paginator[T]) prepareMessageContent(state PaginationState, embed *discordgo.MessageEmbed) (embeds []*discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	embeds = []*discordgo.MessageEmbed{embed}
	components = []discordgo.MessageComponent{
		p.CreatePaginationButtons(state),
	}
	return
}

func (p *Paginator[T]) Setup(session *discordgo.Session, i *discordgo.InteractionCreate, channelID string, extraData map[string]string, createParams ...any) error {
	initialState := PaginationState{
		CurrentPage: 0,
		ExtraData:   extraData,
	}
	
	data, err := p.GetData(initialState)
	if err != nil {
		return fmt.Errorf("error getting initial data: %v", err)
	}
	initialState.TotalPages = p.CalculateTotalPages(len(data))

	pageData, err := p.GetPageData(initialState)
	if err != nil {
		return fmt.Errorf("error getting initial page data: %v", err)
	}

	// if create is nil, default to update (used in lead command)
	var embed *discordgo.MessageEmbed
	if p.Create == nil {
		if len(createParams) != 0 {
			return fmt.Errorf("Paginator.Create is nil, but createParams were provided")
		}

		embed, err = p.Update(initialState, pageData, nil)
	} else {
		embed, err = p.Create(initialState, pageData, createParams...)
	}
	if err != nil {
		return fmt.Errorf("error creating initial pagination embed: %v", err)
	}

	embeds, components := p.prepareMessageContent(initialState, embed)

	interactions.SendMessageComplex(session, i, channelID, "", &components, &embeds, false)
	return nil
}

// edits the message to reflect the new page
// editMessage assumes that there's only one embed in the message
// TODO: avoid panic if there are no embeds
func (p *Paginator[T]) editMessage(s *discordgo.Session, i *discordgo.InteractionCreate, state PaginationState) (err error) {
	data, err := p.GetPageData(state)
	if err != nil {
		return fmt.Errorf("error getting page data: %v", err)
	}

	embed, err := p.Update(state, data, i.Message.Embeds[0])
	if err != nil {
		return
	}

	embeds, components := p.prepareMessageContent(state, embed)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     embeds,
			Components: components,
		},
	})
	return
}

func (p *Paginator[T]) launchJumpModal(s *discordgo.Session, i *discordgo.InteractionCreate, data []string) error {
	state, err := p.GetStateFromCustomId(data)
	if err != nil {
		return err
	}
	
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: p.GetComponentIdWithData(state, JUMP_MODAL),
			Title: "Go to Page",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID: jumpModalInputId,
							Label:    fmt.Sprintf("Enter a page number (1-%d):", state.TotalPages),
							Style: discordgo.TextInputShort,
							Placeholder: "1",
							Required: true,
							MaxLength: 3,
							MinLength: 1,
						},
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Println(util.Fail("Error launching jump to page modal: %v", err))
		interactions.SendEphemeralMessage(s, i, "Error displaying jump to page modal.")
		return fmt.Errorf("error launching jump to page modal: %v", err)
	}
	return nil
}
