package bot

import (
	"fmt"

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

func init() {
	commandHandlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate){
		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			helpcmd(s, nil, i)
		},
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			pingcmd(s, nil, i)
		},
		"roleme": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			data := i.ApplicationCommandData()
			teamID := getStringOption(data.Options, "team_id")
			if teamID == "" {
				sendMessage(s, i, "", "Please provide a team number.")
				return
			}
			rolemeCmd(s, nil, i, []string{teamID})
		},
		"team": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			data := i.ApplicationCommandData()
			var args []string
			if len(data.Options) > 0 {
				sub := data.Options[0]
				subName := sub.Name
				switch subName {
				case "info", "stats", "awards":
					teamID := getStringOption(sub.Options, "team_id")
					if teamID == "" {
						sendMessage(s, i, "", "Please provide a team number.")
						return
					}
					args = []string{teamID, subName}
				default:
					sendMessage(s, i, "", "Unknown subcommand for team.")
					return
				}
			}
			teamcmd(s, nil, i, args)
		},
		"match": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			data := i.ApplicationCommandData()
			if len(data.Options) == 0 {
				sendMessage(s, i, "", "Please provide a subcommand for match.")
				return
			}
			sub := data.Options[0]
			subName := sub.Name
			switch subName {
			case "info":
				year := getStringOption(sub.Options, "year")
				eventCode := getStringOption(sub.Options, "event_code")
				matchNumber := getStringOption(sub.Options, "match_number")
				if year == "" || eventCode == "" || matchNumber == "" {
					sendMessage(s, i, "", "Usage: /match info <year> <event_code> <match_number>")
					return
				}
				matchcmd(s, nil, i, []string{"info", year, eventCode, matchNumber})
			case "eventstart":
				year := getStringOption(sub.Options, "year")
				eventCode := getStringOption(sub.Options, "event_code")
				if year == "" || eventCode == "" {
					sendMessage(s, i, "", "Usage: /match eventstart <year> <event_code>")
					return
				}
				matchcmd(s, nil, i, []string{"eventstart", year, eventCode})
			default:
				sendMessage(s, i, "", "Unknown subcommand for match.")
			}
		},
		"lead": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			data := i.ApplicationCommandData()
			year := getStringOption(data.Options, "year")
			eventCode := getStringOption(data.Options, "event_code")
			if year == "" || eventCode == "" {
				sendMessage(s, i, "", "Usage: /lead <year> <event_code>")
				return
			}
			leadcmd(s, nil, i, []string{year, eventCode})
		},
		"mech": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			data := i.ApplicationCommandData()
			if len(data.Options) == 0 {
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Content: "No subcommand provided.", Flags: 1 << 6}})
				return
			}
			sub := data.Options[0]
			subName := sub.Name

			isAdminUser, err := isAdmin(s, i.GuildID, i.Member.User.ID)
			if err != nil {
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Content: "Unable to check permissions.", Flags: 1 << 6}})
				return
			}
			if !isAdminUser {
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Content: "You do not have permission to run this command.", Flags: 1 << 6}})
				return
			}

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			switch subName {
			case "restart":
				restartBot(s, i.ChannelID, i)
			default:
				sendMessage(s, i, "", "Unknown mech subcommand.")
			}
		},
	}
}

func sendEmbed(session *discordgo.Session, i *discordgo.InteractionCreate, channelID string, embed *discordgo.MessageEmbed) {
	if i != nil {
		_, err := session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}})

		if err != nil {
			msg := fmt.Sprintf("Failed to send embed: %v", err)
			session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		}
	} else {
		_, err := session.ChannelMessageSendEmbed(channelID, embed)

		if err != nil {
			session.ChannelMessageSend(channelID, fmt.Sprintf("Failed to send embed: %v", err))
		}
	}
}

func sendMessage(session *discordgo.Session, i *discordgo.InteractionCreate, channelID string, message string) {
	if i != nil {
		_, err := session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &message})
		// err := session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 	Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 	Data: &discordgo.InteractionResponseData{
		// 		Content: message,
		// 	},
		// })
		if err != nil {
			msg := fmt.Sprintf("Failed to send message: %v", err)
			session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		}
	} else {
		_, err := session.ChannelMessageSend(channelID, message)
		if err != nil {
			session.ChannelMessageSend(channelID, fmt.Sprintf("Failed to send message: %v", err))
		}
	}
}

func getGuildId(message *discordgo.MessageCreate, i *discordgo.InteractionCreate) string {
	if i != nil {
		return i.GuildID
	}
	if message != nil {
		return message.GuildID
	}
	panic("Both message and interaction are nil in getGuildId")
}

func getAuthorId(message *discordgo.MessageCreate, i *discordgo.InteractionCreate) string {
	if i != nil {
		return i.Member.User.ID
	}
	if message != nil {
		return message.Author.ID
	}
	panic("Both message and interaction are nil in getAuthorId")
}

func getChannelId(message *discordgo.MessageCreate, i *discordgo.InteractionCreate) string {
	if i != nil {
		return i.ChannelID
	}
	if message != nil {
		return message.ChannelID
	}
	panic("Both message and interaction are nil in getChannelId")
}
