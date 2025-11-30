package bot

import "github.com/bwmarrin/discordgo"

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
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "team_id",
				Description: "The FTC team ID to look up.",
				Required:    true,
			},
		},
	},
	{
		Name:        "ping",
		Description: "Checks the bot's responsiveness.",
	},
	{
		Name:        "match",
		Description: "Provides information about a specific match.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "match_id",
				Description: "The match ID to look up.",
				Required:    true,
			},
		},
	},
}
