package bot

import (
	"github.com/bwmarrin/discordgo"
)

var cmd_prefix = ">>"

// all the registered slash commands (populated by each command file's init())
var commands []*discordgo.ApplicationCommand

var FtcYearChoices []*discordgo.ApplicationCommandOptionChoice = []*discordgo.ApplicationCommandOptionChoice{
	{Name:  "2025-2026", Value: "2025"},
	{Name:  "2024-2025", Value: "2024"},
	{Name:  "2023-2024", Value: "2023"},
	{Name:  "2022-2023", Value: "2022"},
	{Name:  "2021-2022", Value: "2021"},
	{Name:  "2020-2021", Value: "2020"},
	{Name:  "2019-2020", Value: "2019"},
}

// commandHandlers maps top-level command names to interaction handlers.
var commandHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)

var autocompleteHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)

var componentHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate, string)

func RegisterCommand(cmd *discordgo.ApplicationCommand, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
	if commandHandlers == nil {
		commandHandlers = make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
	}
	commands = append(commands, cmd)
	commandHandlers[cmd.Name] = handler
}

func RegisterAutocompleteHandler(cmdName string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
	if autocompleteHandlers == nil {
		autocompleteHandlers = make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
	}
	autocompleteHandlers[cmdName] = handler
}

func RegisterComponentHandler(customID string, handler func(*discordgo.Session, *discordgo.InteractionCreate, string)) {
	if componentHandlers == nil {
		componentHandlers = make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate, string))
	}
	componentHandlers[customID] = handler

	if len(componentHandlers) >= MAX_COMPONENT_HANDLERS {
		// remove an arbitrary component handler to free up space
		for k := range componentHandlers {
			delete(componentHandlers, k)
			break
		}
	}
}

// utility to get string option from interaction data options
func getStringOption(opts []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, o := range opts {
		if o.Name == name && o.Value != nil {
			if v, ok := o.Value.(string); ok {
				return v
			}
		}
		// if this option is a subcommand, search its children
		if len(o.Options) > 0 {
			if v := getStringOption(o.Options, name); v != "" {
				return v
			}
		}
	}
	return ""
}
