package pagination

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
	"github.com/shuban-789/bjorn/src/bot/util"
)

func (p *Paginator) Register() {
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

func (p *Paginator) pageLeftRight(s *discordgo.Session, ic *discordgo.InteractionCreate, data []string, delta int) error {
	state, err := p.GetStateFromCustomId(data)
	if err != nil {
		return err
	}
	
	state.CurrentPage = util.Clamp(state.CurrentPage+delta, 0, state.TotalPages-1)
	return p.editMessage(s, ic, state)
}


func (p *Paginator) handleJumpModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate, id_data []string, modal_data discordgo.ModalSubmitInteractionData) error {
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

func (p *Paginator) prepareMessageContent(state PaginationState, embed *discordgo.MessageEmbed) (embeds []*discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	embeds = []*discordgo.MessageEmbed{embed}
	components = []discordgo.MessageComponent{
		p.CreatePaginationButtons(state),
	}
	return
}

func (p *Paginator) Setup(session *discordgo.Session, i *discordgo.InteractionCreate, channelID string, initialState PaginationState, createParams ...any) error {
	embed, err := p.Create(initialState, createParams...)
	if err != nil {
		return fmt.Errorf("error creating initial pagination embed: %v", err)
	}

	embeds, components := p.prepareMessageContent(initialState, embed)

	interactions.SendMessageComplex(session, i, channelID, "", &components, &embeds, false)
	return nil
}

// edits the message to reflect the new page
// editMessage assumes that there's only one embed in the message
func (p *Paginator) editMessage(s *discordgo.Session, i *discordgo.InteractionCreate, state PaginationState) (err error) {
	embed, err := p.Update(state, i.Message.Embeds[0])
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

func (p *Paginator) launchJumpModal(s *discordgo.Session, i *discordgo.InteractionCreate, data []string) error {
	state, err := p.GetStateFromCustomId(data)
	if err != nil {
		return err
	}
	
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf(p.GetComponentIdWithData(state, JUMP_MODAL)),
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
