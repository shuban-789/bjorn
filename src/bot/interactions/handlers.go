package interactions

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/util"
)

var CmdPrefix = ">>"

// all the registered slash commands (populated by each command file's init())
var Commands []*discordgo.ApplicationCommand

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
var CommandHandlers map[string]CommandHandler

// map of top level cmds for custom autocomplete handlers so you can write more custom code if needed
// basically this is just the old system but I didn't want to delete it just in case
var AutocompleteHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)

// func that returns a list of autocomplete choices given the current options
// opts is the current options filled out (like the region in the match track command is used to get the list of events)
// query is the current value of the focused option being typed
type AutocompleteProvider func(opts map[string]string, query string) []*discordgo.ApplicationCommandOptionChoice

// maps str of "command/subcommand/option" or "command/option" to the right function
var AutocompleteProviders map[string]AutocompleteProvider

// maps custom ID of component to handler func
type ComponentHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, data string)
var ComponentHandlers map[string]ComponentHandler

// maps custom ID of modal to handler func
type ModalHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, id_data string, modal_data discordgo.ModalSubmitInteractionData)
var ModalHandlers map[string]ModalHandler

func RegisterCommand(cmd *discordgo.ApplicationCommand, handler CommandHandler) {
	if CommandHandlers == nil {
		CommandHandlers = make(map[string]CommandHandler)
	}
	Commands = append(Commands, cmd)
	CommandHandlers[cmd.Name] = handler
}

// this is if you want to register a custom autocomplete handler for a command name, but it's better to use registerautocomplete now
// I just kept this just in case
func RegisterAutocompleteHandlerCustom(cmdName string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
	if AutocompleteHandlers == nil {
		AutocompleteHandlers = make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
	}
	AutocompleteHandlers[cmdName] = handler
}

// simpler autocomplete mapping system for a specific command/subcommand/option path.
// path format: "command/subcommand/option" or "command/option" (for commands without subcommands)
// provider func returns list of choices
func RegisterAutocomplete(path string, provider AutocompleteProvider) {
	if AutocompleteProviders == nil {
		AutocompleteProviders = make(map[string]AutocompleteProvider)
	}
	AutocompleteProviders[path] = provider
}

func HandleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	provider, ok := AutocompleteProviders[path]
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
	if ComponentHandlers == nil {
		ComponentHandlers = make(map[string]ComponentHandler)
	}
	ComponentHandlers[customID] = handler
}

func RegisterModalHandler(customID string, handler ModalHandler) {
	if ModalHandlers == nil {
		ModalHandlers = make(map[string]ModalHandler)
	}
	ModalHandlers[customID] = handler
}

// utility to get string option from interaction data options
func GetStringOption(opts []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, o := range opts {
		if o.Name == name && o.Value != nil {
			if v, ok := o.Value.(string); ok {
				return v
			}
		}
		// if this option is a subcommand, search its children
		if len(o.Options) > 0 {
			if v := GetStringOption(o.Options, name); v != "" {
				return v
			}
		}
	}
	return ""
}
