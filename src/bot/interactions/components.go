// This file contains functions to use components (buttons, selects) in interactions.
// I've only implemented buttons so far though since we haven't really needed anything else yet.

package interactions

import (
	"github.com/bwmarrin/discordgo"
)

// this might be redundant since I just added it into interop.go, I'm thinking of turning this into a class for registering component handlers later

func GetComponentWithId(components []discordgo.MessageComponent, id string) discordgo.MessageComponent {
	for _, comp := range components {
		compType := comp.Type()
		switch compType {
		case discordgo.ActionsRowComponent:
			ar := comp.(*discordgo.ActionsRow)
			foundComp := GetComponentWithId(ar.Components, id)
			if foundComp != nil {
				return foundComp
			}
		case discordgo.ButtonComponent:
			btn := comp.(*discordgo.Button)
			if btn.CustomID == id {
				return btn
			}
		case discordgo.TextInputComponent:
			ti := comp.(*discordgo.TextInput)
			if ti.CustomID == id {
				return ti
			}
		case discordgo.SelectMenuComponent, discordgo.UserSelectMenuComponent, 
			discordgo.RoleSelectMenuComponent, discordgo.MentionableSelectMenuComponent, 
			discordgo.ChannelSelectMenuComponent:
			sel := comp.(*discordgo.SelectMenu)
			if sel.CustomID == id {
				return sel
			}
		}
	}
	return nil
}
