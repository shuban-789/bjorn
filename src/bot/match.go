package bot

import (
	"fmt"
	"strings"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

func matchcmd(channelID string, args []string, session *discordgo.Session, guildId string, authorID string) {
    if len(args) < 1 {
        session.ChannelMessageSend(channelID, "Please provide a subcommand (e.g., 'info').")
        return
    }

    subCommand := args[0]

    switch subCommand {
    case "info":
        if len(args) < 4 {
            session.ChannelMessageSend(channelID, "Usage: `>>match info <year> <eventCode> <matchNumber>`")
            return
        }

        year := args[1]
        eventCode := args[2]
        matchNumber := args[3]

        getMatch(channelID, year, eventCode, matchNumber, session)
	case "eventstart":
		if len(args) < 3 {
			session.ChannelMessageSend(channelID, "Usage: `>>match eventstart <year> <eventCode>`")
			return
		}

		if isAdmin(session, guildId, authorID) {
			session.ChannelMessageSend(channelID, "You do not have permission to run this command.")
			return
		}

		year := args[1]
		eventCode := args[2]
		eventStart(channelID, year, eventCode, session)
    default:
        session.ChannelMessageSend(channelID, "Unknown subcommand. Available subcommands: `info`")
    }
}

func eventStart(channelID string, args []string, session *discordgo.Session) {
	if len(args) < 3 {
		session.ChannelMessageSend(channelID, "Usage: `>>e start <year> <eventCode>`")
		return
	}

	year := args[1]
	eventCode := args[2]

	getEventStart(channelID, year, eventCode, session)
}

func getMatch(ChannelID string, year string, eventCode string, matchNumber string, session *discordgo.Session) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s/matches", year, eventCode)
	resp, err := http.Get(url)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to fetch match data: %v", err))
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to read response: %v", err))
		return
	}

	var matches []struct {
		ID     int `json:"id"`
		Scores struct {
			Red struct {
				TotalPoints int `json:"totalPoints"`
			} `json:"red"`
			Blue struct {
				TotalPoints int `json:"totalPoints"`
			} `json:"blue"`
		} `json:"scores"`
		Teams []struct {
			Alliance   string `json:"alliance"`
			TeamNumber int    `json:"teamNumber"`
		} `json:"teams"`
	}
	err = json.Unmarshal(body, &matches)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to parse match data: %v", err))
		return
	}

	var selectedMatch struct {
		Scores struct {
			Red  int `json:"totalPoints"`
			Blue int `json:"totalPoints"`
		}
		RedTeams  []int
		BlueTeams []int
	}
	for _, match := range matches {
		if fmt.Sprintf("%d", match.ID) == matchNumber {
			selectedMatch.Scores.Red = match.Scores.Red.TotalPoints
			selectedMatch.Scores.Blue = match.Scores.Blue.TotalPoints
			for _, team := range match.Teams {
				if team.Alliance == "Red" {
					selectedMatch.RedTeams = append(selectedMatch.RedTeams, team.TeamNumber)
				} else if team.Alliance == "Blue" {
					selectedMatch.BlueTeams = append(selectedMatch.BlueTeams, team.TeamNumber)
				}
			}
			break
		}
	}

	winner := "Red Alliance"
	if selectedMatch.Scores.Blue > selectedMatch.Scores.Red {
		winner = "Blue Alliance"
	} else if selectedMatch.Scores.Blue == selectedMatch.Scores.Red {
		winner = "Draw"
	}

	color := 0xE02C44
	if strings.Compare(winner, "Blue Alliance") == 0 {
		color = 0x58ACEC
	} else if strings.Compare(winner, "Draw") == 0 {
		color = 0xE8E4EC
	}
	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s Match %s Results", eventCode, matchNumber),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Red Alliance Teams  🔴   \u200B",
				Value:  fmt.Sprintf("%v", selectedMatch.RedTeams),
				Inline: true,
			},
			{
				Name:   "Blue Alliance Teams  🔵   \u200B",
				Value:  fmt.Sprintf("%v", selectedMatch.BlueTeams),
				Inline: true,
			},
			{
				Name:   "\u200b",
				Value:  "",
				Inline: false,
			},
			{
				Name:   "Red Alliance Score  🔴   \u200B",
				Value:  fmt.Sprintf("** **%d", selectedMatch.Scores.Red),
				Inline: true,
			},
			{
				Name:   "Blue Alliance Score  🔵   \u200B",
				Value:  fmt.Sprintf("%d", selectedMatch.Scores.Blue),
				Inline: true,
			},
			{
				Name:   "\u200b",
				Value:  "",
				Inline: false,
			},
			{
				Name:   "Winner  🏆\u200B",
				Value:  winner,
				Inline: false,
			},
		},
		Color: color,
	}
	
	
	session.ChannelMessageSendEmbed(ChannelID, embed)
}
