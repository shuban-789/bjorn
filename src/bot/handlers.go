package bot

import (
	"github.com/bwmarrin/discordgo"
)

var cmd_prefix = ">>"

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "help",
		Description: "Displays help information about the bot commands.",
	},
	{
		Name:        "roleme",
		Description: "Assigns you a role based on your team ID.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "team_id",
				Description: "Your FTC team ID.",
				Required:    true,
			},
		},
	},
	{
		Name:        "team",
		Description: "Provides information about a specific FTC team.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "info",
				Description: "Show general team information.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "team_id",
						Description: "The FTC team ID to look up.",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "stats",
				Description: "Show team statistics.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "team_id",
						Description: "The FTC team ID to look up.",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "awards",
				Description: "Show awards for a team.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "team_id",
						Description: "The FTC team ID to look up.",
						Required:    true,
					},
				},
			},
		},
	},
	{
		Name:        "ping",
		Description: "Checks the bot's responsiveness.",
	},
	{
		Name:        "match",
		Description: "Provides information and controls for matches.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "info",
				Description: "Lookup information about a certain match.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "year",
						Description: "Year of the event (e.g., 2025).",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "event_code",
						Description: "The event code to look up.",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "match_number",
						Description: "The match ID/number to look up.",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "eventstart",
				Description: "Start an active match tracker for a current event.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "year",
						Description: "Year of the event (e.g., 2025).",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "event_code",
						Description: "The event code to track.",
						Required:    true,
					},
				},
			},
		},
	},
	{
		Name:        "lead",
		Description: "Display the leaderboard for a certain event.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "year",
				Description: "Year of the event (e.g., 2025).",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "event_code",
				Description: "Event code to look up.",
				Required:    true,
			},
		},
	},
	{
		Name:                     "mech",
		Description:              "Mechanic/admin commands for the bot.",
		DefaultMemberPermissions: func() *int64 { p := int64(discordgo.PermissionAdministrator); return &p }(),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "restart",
				Description: "Restart the bot.",
			},
		},
	},
}

// helper to read string options safely
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

// commandHandlers maps top-level command names to interaction handlers.
var commandHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)

var componentHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)

func RegisterCommand(name string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
	if commandHandlers == nil {
		commandHandlers = make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
	}
	commandHandlers[name] = handler
}

func RegisterComponentHandler(customID string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
	if componentHandlers == nil {
		componentHandlers = make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate))
	}
	componentHandlers[customID] = handler
}
