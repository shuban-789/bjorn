package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/util"
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
type CommandHandler func(*discordgo.Session, *discordgo.InteractionCreate)
var commandHandlers map[string]CommandHandler

// map of top level cmds for custom autocomplete handlers so you can write more custom code if needed
// basically this is just the old system but I didn't want to delete it just in case
var autocompleteHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)

// func that returns a list of autocomplete choices given the current options
// opts is the current options filled out (like the region in the match track command is used to get the list of events)
// query is the current value of the focused option being typed
type AutocompleteProvider func(opts map[string]string, query string) []*discordgo.ApplicationCommandOptionChoice

// maps str of "command/subcommand/option" or "command/option" to the right function
var autocompleteProviders map[string]AutocompleteProvider

// maps custom ID of component to handler func
type ComponentHandler func(*discordgo.Session, *discordgo.InteractionCreate, string)
var componentHandlers map[string]ComponentHandler

func RegisterCommand(cmd *discordgo.ApplicationCommand, handler CommandHandler) {
	if commandHandlers == nil {
		commandHandlers = make(map[string]CommandHandler)
	}
	commands = append(commands, cmd)
	commandHandlers[cmd.Name] = handler
}

// this is if you want to register a custom autocomplete handler for a command name, but it's better to use registerautocomplete now
// I just kept this just in case
func RegisterAutocompleteHandlerCustom(cmdName string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
	if autocompleteHandlers == nil {
		autocompleteHandlers = make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
	}
	autocompleteHandlers[cmdName] = handler
}

// simpler autocomplete mapping system for a specific command/subcommand/option path.
// path format: "command/subcommand/option" or "command/option" (for commands without subcommands)
// provider func returns list of choices
func RegisterAutocomplete(path string, provider AutocompleteProvider) {
	if autocompleteProviders == nil {
		autocompleteProviders = make(map[string]AutocompleteProvider)
	}
	autocompleteProviders[path] = provider
}

func handleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	cmdName := data.Name

	if len(data.Options) == 0 {
		return
	}

	// check for if the first one is actually a subcommand
	firstOpt := data.Options[0]
	var subName string
	var options []*discordgo.ApplicationCommandInteractionDataOption

	if firstOpt.Type == discordgo.ApplicationCommandOptionSubCommand ||
		firstOpt.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
		subName = firstOpt.Name
		options = firstOpt.Options
	} else {
		subName = ""
		options = data.Options
	}

	var focusedOpt *discordgo.ApplicationCommandInteractionDataOption
	opts := make(map[string]string)

	for _, opt := range options {
		if opt.Value != nil {
			if v, ok := opt.Value.(string); ok {
				opts[opt.Name] = v
			}
		}
		if opt.Focused {
			focusedOpt = opt
		}
	}

	if focusedOpt == nil {
		return
	}

	// this is the key we index in the map
	var path string
	if subName != "" {
		path = cmdName + "/" + subName + "/" + focusedOpt.Name
	} else {
		path = cmdName + "/" + focusedOpt.Name
	}

	provider, ok := autocompleteProviders[path]
	if !ok {
		return
	}

	query := ""
	if focusedOpt.Value != nil {
		if v, ok := focusedOpt.Value.(string); ok {
			query = v
		}
	}

	choices := provider(opts, query)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})

	if err != nil {
		fmt.Println(util.Fail("Error responding to autocomplete interaction: %v", err))
	}
}

func RegisterComponentHandler(customID string, handler ComponentHandler) {
	if componentHandlers == nil {
		componentHandlers = make(map[string]ComponentHandler)
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
