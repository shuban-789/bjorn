package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
)

func init() {
	RegisterCommand(
		&discordgo.ApplicationCommand{
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
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
			data := i.ApplicationCommandData()
			if len(data.Options) == 0 {
				interactions.SendMessage(s, i, "", "Please provide a subcommand for match.")
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
					interactions.SendMessage(s, i, "", "Usage: /match info <year> <event_code> <match_number>")
					return
				}
				matchcmd(s, nil, i, []string{"info", year, eventCode, matchNumber})
			case "eventstart":
				year := getStringOption(sub.Options, "year")
				eventCode := getStringOption(sub.Options, "event_code")
				if year == "" || eventCode == "" {
					interactions.SendMessage(s, i, "", "Usage: /match eventstart <year> <event_code>")
					return
				}
				matchcmd(s, nil, i, []string{"eventstart", year, eventCode})
			default:
				interactions.SendMessage(s, i, "", "Unknown subcommand for match.")
			}
		},
	)
}

type AllianceColor int

const (
	Red AllianceColor = iota
	Blue
	Neither // in the case of a draw in match results
)

type TwoTeamAlliance struct {
	Captain   int
	FirstPick int
	Color     AllianceColor
}

type DoubleElimTournament struct {
	WinnersBracket map[TwoTeamAlliance]int  // alliance to current match number
	LosersBracket  map[TwoTeamAlliance]int  // alliance to current match number
	Eliminated     map[TwoTeamAlliance]bool // teams that are eliminated
	MatchHistory   []MatchResult            // history of match results for visualization
}

type Match struct {
	ID              int    `json:"id"`
	HasBeenPlayed   bool   `json:"hasBeenPlayed"`
	ActualStartTime string `json:"actualStartTime"`
	TournamentLevel string `json:"tournamentLevel"`
}

func (m *Match) GetHasBeenPlayed() bool {
	return m.HasBeenPlayed
}

type MatchResult struct {
	RedAlliance  TwoTeamAlliance
	BlueAlliance TwoTeamAlliance
	RedScore     int
	BlueScore    int
	Winner       AllianceColor
	MatchNumber  string
}

type EventTracked struct {
	Year                 string
	EventCode            string
	UpdateChannelId      string
	LastUpdateTime       time.Time
	CachedMatches        []Match
	LastProcessedMatchId int
	Ongoing              bool
	StartTime            time.Time
	EndTime              time.Time
	Tournament           DoubleElimTournament
}

var eventsBeingTracked []EventTracked

type EventDetails struct {
	Name          string `json:"name"`
	Start         string `json:"start"`
	End           string `json:"end"`
	Venue         string `json:"venue"`
	Address       string `json:"address"`
	Country       string `json:"country"`
	City          string `json:"city"`
	State         string `json:"state"`
	Website       string `json:"website"`
	LiveStreamUrl string `json:"liveStreamUrl"`
	Ongoing       bool   `json:"ongoing"`
	TimeZone      string `json:"timezone"`
}

// this is used in the api call to get a match, it's a small part of it but I use this in other funcs so I define it globally
type TeamDTO struct {
	AllianceColor string `json:"alliance"`
	AllianceRole  string `json:"allianceRole"`
	TeamNumber    int    `json:"teamNumber"`
}

func matchcmd(session *discordgo.Session, message *discordgo.MessageCreate, i *discordgo.InteractionCreate, args []string) {
	authorId := interactions.GetAuthorId(message, i)
	guildId := interactions.GetGuildId(message, i)
	channelId := interactions.GetChannelId(message, i)

	if len(args) < 1 {
		interactions.SendMessage(session, i, channelId, "Please provide a subcommand (e.g., 'info').")
		return
	}

	subCommand := args[0]

	switch subCommand {
	case "info":
		if len(args) < 4 {
			interactions.SendMessage(session, i, channelId, "Usage: `>>match info <year> <eventCode> <matchNumber>`")
			return
		}

		year := args[1]
		eventCode := args[2]
		matchNumber := args[3]

		getMatch(channelId, year, eventCode, matchNumber, nil, session)
	case "eventstart":
		if len(args) < 3 {
			interactions.SendMessage(session, i, channelId, "Usage: `>>match eventstart <year> <eventCode>`")
			return
		}

		hasPerms, err := isAdmin(session, guildId, authorId)
		if err != nil {
			interactions.SendMessage(session, i, channelId, "Unable to check permissions of user.")
			return
		}

		if hasPerms {
			interactions.SendMessage(session, i, channelId, "You do not have permission to run this command.")
			return
		}

		year := args[1]
		eventCode := args[2]
		eventStart(channelId, guildId, year, eventCode, session, i)
	default:
		interactions.SendMessage(session, i, channelId, "Unknown subcommand. Available subcommands: `info`, `eventstart`")
	}
}

func eventStart(channelID, guildID, year, eventCode string, session *discordgo.Session, i *discordgo.InteractionCreate) {
	eventDetails, err := fetchEventDetails(year, eventCode)
	if err != nil {
		interactions.SendMessage(session, i, channelID, err.Error())
		return
	}

	location, err := time.LoadLocation(eventDetails.TimeZone)
	if err != nil {
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Error loading event timezone: %v", err))
		return
	}

	today := time.Now().In(location)
	startTime, endTime, err := getEventStartEndTime(eventDetails, today, location)
	if err != nil {
		interactions.SendMessage(session, i, channelID, err.Error())
		return
	}

	if endTime.Before(today) {
		interactions.SendMessage(session, i, channelID, "This event has already ended!")
		return
	}

	description := fmt.Sprintf("Event: %s\nVenue: %s\nAddress: %s, %s, %s, %s\nWebsite: %s\nLive Stream: %s",
		eventDetails.Name, eventDetails.Venue, eventDetails.Address, eventDetails.City, eventDetails.State, eventDetails.Country, eventDetails.Website, eventDetails.LiveStreamUrl)

	eventsBeingTracked = append(eventsBeingTracked, EventTracked{
		Year:                 year,
		EventCode:            eventCode,
		UpdateChannelId:      channelID,
		LastUpdateTime:       time.Date(1, time.January, 1, 1, 1, 1, 1, time.Now().Location()), // hopefully will force an immediate update
		CachedMatches:        []Match{},
		LastProcessedMatchId: -100, // should hopefully force update
		Ongoing:              eventDetails.Ongoing,
		StartTime:            startTime,
		EndTime:              endTime,
		Tournament: DoubleElimTournament{
			WinnersBracket: make(map[TwoTeamAlliance]int),
			LosersBracket:  make(map[TwoTeamAlliance]int),
			Eliminated:     make(map[TwoTeamAlliance]bool),
		},
	})
	interactions.SendMessage(session, i, channelID, fmt.Sprintf("Started tracking matches for event %s in %s...", eventCode, year))

	if startTime.Before(today) {
		interactions.SendMessage(session, i, channelID, "This event has already started! Will not create event.")
		return
	}

	// create the discord event
	event, err := session.GuildScheduledEventCreate(guildID, &discordgo.GuildScheduledEventParams{
		Name:               eventDetails.Name,
		Description:        description,
		ScheduledStartTime: &startTime,
		ScheduledEndTime:   &endTime,
		PrivacyLevel:       discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
		EntityType:         discordgo.GuildScheduledEventEntityTypeExternal, // stage, voice, or external I think
		EntityMetadata: &discordgo.GuildScheduledEventEntityMetadata{
			Location: eventDetails.Venue,
		},
	})
	if err != nil {
		log.Fatalf("Error creating event: %v", err)
	}
	interactions.SendMessage(session, i, channelID, fmt.Sprintf("Created event: %s", event.ID))
}

func getMatch(ChannelID string, year string, eventCode string, matchNumber string, event *EventTracked, session *discordgo.Session) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s/matches", year, eventCode)
	resp, err := http.Get(url)
	if err != nil {
		interactions.SendMessage(session, nil, ChannelID, fmt.Sprintf("Failed to fetch match data: %v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		interactions.SendMessage(session, nil, ChannelID, fmt.Sprintf("Failed to read response: %v", err))
		return
	}

	type TeamScoreDetail struct {
		Total  int `json:"totalPoints"`
		Auto   int `json:"autoPoints"`
		TeleOp int `json:"dcPoints"` // "driver controlled"
		Fouls  int `json:"penaltyPointsByOpp"`
	}

	var matches []struct {
		ID     int `json:"id"`
		Scores struct {
			Red  TeamScoreDetail `json:"red"`
			Blue TeamScoreDetail `json:"blue"`
		} `json:"scores"`
		Teams           []TeamDTO `json:"teams"`
		TournamentLevel string    `json:"tournamentLevel"`
		Series          int       `json:"series"`
	}
	err = json.Unmarshal(body, &matches)
	if err != nil {
		interactions.SendMessage(session, nil, ChannelID, fmt.Sprintf("Failed to parse match data: %v", err))
		return
	}

	var selectedMatch struct {
		Scores struct {
			Red  TeamScoreDetail
			Blue TeamScoreDetail
		}
		RedAlliance     TwoTeamAlliance
		RedTeams        []TeamDTO
		BlueAlliance    TwoTeamAlliance
		BlueTeams       []TeamDTO
		TournamentLevel string
		ID              int
		Series          int
	}
	redTeams := []TeamDTO{}
	blueTeams := []TeamDTO{}
	for _, match := range matches {
		if fmt.Sprintf("%d", match.ID) == matchNumber {
			selectedMatch.Scores.Red = match.Scores.Red
			selectedMatch.Scores.Blue = match.Scores.Blue
			for _, team := range match.Teams {
				if team.AllianceColor == "Red" {
					redTeams = append(redTeams, team)
				} else if team.AllianceColor == "Blue" {
					blueTeams = append(blueTeams, team)
				} else {
					// TODO: I'm going to crash out if this happens
				}
				selectedMatch.TournamentLevel = match.TournamentLevel
				selectedMatch.ID = match.ID
				selectedMatch.Series = match.Series
			}
			break
		}
	}
	selectedMatch.RedTeams = redTeams
	selectedMatch.BlueTeams = blueTeams
	selectedMatch.RedAlliance = getAllianceFromTeams(redTeams)
	selectedMatch.BlueAlliance = getAllianceFromTeams(blueTeams)

	winnerSkib := Red
	if selectedMatch.Scores.Blue.Total > selectedMatch.Scores.Red.Total {
		winnerSkib = Blue
	} else if selectedMatch.Scores.Blue.Total == selectedMatch.Scores.Red.Total {
		winnerSkib = Neither
	}

	color := 0xE02C44
	if winnerSkib == Blue {
		color = 0x58ACEC
	} else if winnerSkib == Neither {
		color = 0xE8E4EC
	}

	var matchName string
	if selectedMatch.TournamentLevel == "Quals" {
		matchName = fmt.Sprintf("Qualification %d", selectedMatch.ID)
	} else if selectedMatch.TournamentLevel == "DoubleElim" {
		matchName = fmt.Sprintf("Playoffs Match %d", selectedMatch.Series)
	} else {
		matchName = fmt.Sprintf("Match %d?", selectedMatch.ID)
	}
	useQualsTeamNaming := selectedMatch.TournamentLevel == "Quals"
	var redAlliance strings.Builder
	var blueAlliance strings.Builder
	// supports if there are any number of teams (including more than 2 for some reason)
	nTeamsRed := len(selectedMatch.RedTeams)
	nTeamsBlue := len(selectedMatch.BlueTeams)
	if useQualsTeamNaming {
		for i, team := range selectedMatch.RedTeams {
			if i > 0 {
				if nTeamsRed > 2 {
					redAlliance.WriteString(",")
				}
				if i == (nTeamsRed - 1) {
					redAlliance.WriteString(" and ")
				} else {
					redAlliance.WriteString(" ")
				}
			}
			redAlliance.WriteString(fmt.Sprintf("%d", team.TeamNumber))
		}
		for i, team := range selectedMatch.BlueTeams {
			if i > 0 {
				if nTeamsBlue > 2 {
					blueAlliance.WriteString(",")
				}
				if i == (nTeamsBlue - 1) {
					blueAlliance.WriteString(" and ")
				} else {
					blueAlliance.WriteString(" ")
				}
			}
			blueAlliance.WriteString(fmt.Sprintf("%d", team.TeamNumber))
		}
	} else {
		for i, team := range selectedMatch.RedTeams {
			if i > 0 {
				redAlliance.WriteRune('\n')
			}
			redAlliance.WriteString(fmt.Sprintf("%s: %d", team.AllianceRole, team.TeamNumber))
		}
		for i, team := range selectedMatch.BlueTeams {
			if i > 0 {
				blueAlliance.WriteRune('\n')
			}
			blueAlliance.WriteString(fmt.Sprintf("%s: %d", team.AllianceRole, team.TeamNumber))
		}
	}

	redAlliance.WriteString("\n\n")
	blueAlliance.WriteString("\n\n")

	redAlliance.WriteString(fmt.Sprintf(
		"**%d points",
		selectedMatch.Scores.Red.Total,
	))
	if winnerSkib == Red {
		redAlliance.WriteString(" 🏆")
	}
	redAlliance.WriteString(fmt.Sprintf(
		"**\n • Auto: **%d**\n • TeleOp: **%d**\n • Fouls: **%d**",
		selectedMatch.Scores.Red.Auto,
		selectedMatch.Scores.Red.TeleOp,
		selectedMatch.Scores.Red.Fouls,
	))

	blueAlliance.WriteString(fmt.Sprintf(
		"**%d points",
		selectedMatch.Scores.Blue.Total,
	))
	if winnerSkib == Blue {
		blueAlliance.WriteString(" 🏆")
	}
	blueAlliance.WriteString(fmt.Sprintf(
		"**\n • Auto: **%d**\n • TeleOp: **%d**\n • Fouls: **%d**",
		selectedMatch.Scores.Blue.Auto,
		selectedMatch.Scores.Blue.TeleOp,
		selectedMatch.Scores.Blue.Fouls,
	))

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s %s: Results", eventCode, matchName),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Red Alliance  🔴\u200B",
				Value:  fmt.Sprintf("%v", redAlliance.String()),
				Inline: true,
			},
			{
				Name:   "Blue Alliance  🔵\u200B",
				Value:  fmt.Sprintf("%v", blueAlliance.String()),
				Inline: true,
			},
		},
		Color: color,
	}

	discordMsg := &discordgo.MessageSend{
		Embed: embed,
	}

	// // I will not handle the case of a draw because complexity & they are so insanely rare if not impossible because of how many tiebreakers there are
	// if event != nil && selectedMatch.TournamentLevel == "DoubleElim" {
	// 	result := MatchResult{
	// 		RedAlliance:  selectedMatch.RedAlliance,
	// 		BlueAlliance: selectedMatch.BlueAlliance,
	// 		RedScore:     selectedMatch.Scores.Red,
	// 		BlueScore:    selectedMatch.Scores.Blue,
	// 		Winner:       Red,
	// 		MatchNumber:  matchNumber,
	// 	}
	// 	event.Tournament.MatchHistory = append(event.Tournament.MatchHistory, result)

	// 	updateBracket(result, &event.Tournament)
	// 	updateBracket(result, &event.Tournament)

	// 	image := createBracketImage(&event.Tournament)

	// 	var buf bytes.Buffer
	// 	err := png.Encode(&buf, image)
	// 	if err != nil {
	// 		interactions.SendMessage(session, nil, ChannelID, fmt.Sprintf("Failed to encode image: %v", err))
	// 		return
	// 	}

	// 	discordMsg.Files = []*discordgo.File{
	// 		{
	// 			Name:   "bracket_image.png", // The file name as it will appear in Discord
	// 			Reader: &buf,
	// 		},
	// 	}
	// }

	_, err = session.ChannelMessageSendComplex(ChannelID, discordMsg)
	HandleErr(err)
}

func eventUpdate(apiPollTime time.Duration, session *discordgo.Session) {
	for i := 0; i < len(eventsBeingTracked); i++ {
		event := &eventsBeingTracked[i]

		fmt.Print(info("Checking event: Year=%s, EventCode=%s, LastUpdateTime=%v\n",
			event.Year, event.EventCode, event.LastUpdateTime))

		if time.Since(event.LastUpdateTime) < apiPollTime {
			fmt.Print(info("Skipping update for event %s/%s (last updated %v)\n",
				event.Year, event.EventCode, event.LastUpdateTime))
			continue
		}

		event.LastUpdateTime = time.Now()

		eventDetails, err := fetchEventDetails(event.Year, event.EventCode)
		if err != nil {
			session.ChannelMessageSend(event.UpdateChannelId, err.Error())
			return
		}
		if eventDetails.Ongoing && !event.Ongoing {
			event.Ongoing = true
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("The %s has started!", eventDetails.Name))
		} else if !eventDetails.Ongoing && event.Ongoing {
			// remove the event once it's done
			eventsBeingTracked = append(eventsBeingTracked[:i], eventsBeingTracked[i+1:]...)
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("The %s has ended!", eventDetails.Name))
			return
		}

		url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s/matches", event.Year, event.EventCode)
		fmt.Print(info("Fetching matches from: %s\n", url))

		resp, err := http.Get(url)
		if err != nil {
			fmt.Print(fail("Failed to fetch match data for event %s/%s: %v\n",
				event.Year, event.EventCode, err))
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Failed to fetch match data: %v", err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			fmt.Print(info("Event %s/%s not found!\n", event.Year, event.EventCode))
			session.ChannelMessageSend(event.UpdateChannelId, "That event does not exist!")
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Print(fail("Failed to read response for event %s/%s: %v\n",
				event.Year, event.EventCode, err))
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Failed to read response: %v", err))
			continue
		}

		var matches []Match
		err = json.Unmarshal(body, &matches)
		if err != nil {
			fmt.Printf("\033[31m[ERROR]\033[0m Failed to parse match data for event %s/%s: %v\n",
				event.Year, event.EventCode, err)
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Failed to parse match data: %v", err))
			continue
		}

		fmt.Print(info("Matches fetched for event %s/%s: %d matches found\n",
			event.Year, event.EventCode, len(matches)))

		var newMatches []Match
		for _, match := range matches {
			fmt.Print(info("Checking match: ID=%d, hasBeenPlayed=%v\n",
				match.ID, match.GetHasBeenPlayed()))

			if match.ID > event.LastProcessedMatchId && match.GetHasBeenPlayed() {
				newMatches = append(newMatches, match)
				fmt.Print(info("New match to process: ID=%d\n", match.ID))
			}
		}

		if len(newMatches) > 0 {
			for _, match := range newMatches {
				fmt.Print(info("Processing match: ID=%d\n", match.ID))
				getMatch(event.UpdateChannelId, event.Year, event.EventCode, fmt.Sprintf("%d", match.ID), event, session)
			}
			event.LastProcessedMatchId = newMatches[len(newMatches)-1].ID
			fmt.Print(info("Updated LastProcessedMatchId for event %s/%s: %d\n",
				event.Year, event.EventCode, event.LastProcessedMatchId))
		} else {
			fmt.Println(info("No new played matches found in this interval."))
		}
	}
}

func startEventUpdater(session *discordgo.Session, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				eventUpdate(interval, session)
			}
		}
	}()
}

func fetchEventDetails(year, eventCode string) (EventDetails, error) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s", year, eventCode)
	resp, err := http.Get(url)
	if err != nil {
		return EventDetails{}, fmt.Errorf("failed to fetch match data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return EventDetails{}, fmt.Errorf("that event does not exist!")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return EventDetails{}, fmt.Errorf("failed to read response: %v", err)
	}

	var eventDetails EventDetails
	err = json.Unmarshal(body, &eventDetails)
	if err != nil {
		return EventDetails{}, fmt.Errorf("failed to parse event details: %v", err)
	}

	return eventDetails, nil
}

func getEventStartEndTime(eventDetails EventDetails, today time.Time, location *time.Location) (time.Time, time.Time, error) {
	layout := "2006-01-02"
	startTime, err := time.Parse(layout, eventDetails.Start)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse event start time: %v", err)
	}

	endTime, err := time.Parse(layout, eventDetails.End)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse event end time: %v", err)
	}

	startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 8, 0, 0, 0, location)
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 17, 0, 0, 0, location)

	// If we start after 8, set it to be scheduled in the future so it doesn't error out
	if startTime.Year() == today.Year() && startTime.YearDay() == today.YearDay() && today.Hour() >= 8 {
		startTime = today.Add(5 * time.Minute)
	}

	return startTime, endTime, nil
}

func getAllianceFromTeams(teams []TeamDTO) TwoTeamAlliance {
	if len(teams) < 2 {
		return TwoTeamAlliance{}
	}

	var color AllianceColor
	if teams[0].AllianceColor == "Red" {
		color = Red
	} else {
		color = Blue
	}

	if teams[0].AllianceRole == "Captain" {
		return TwoTeamAlliance{Captain: teams[0].TeamNumber, FirstPick: teams[1].TeamNumber, Color: color}
	} else {
		return TwoTeamAlliance{Captain: teams[1].TeamNumber, FirstPick: teams[0].TeamNumber, Color: color}
	}
}

func updateBracket(result MatchResult, tournament *DoubleElimTournament) {
	var winner TwoTeamAlliance
	var loser TwoTeamAlliance
	if result.Winner == Red {
		winner = result.RedAlliance
		loser = result.BlueAlliance
	} else if result.Winner == Blue {
		winner = result.BlueAlliance
		loser = result.RedAlliance
	}

	if _, exists := tournament.Eliminated[loser]; !exists {
		// Loser goes to the losers' bracket
		tournament.LosersBracket[loser] = tournament.LosersBracket[loser] + 1
	} else {
		// Loser is eliminated
		tournament.Eliminated[loser] = true
	}
	// Winner moves forward in the winners' bracket
	tournament.WinnersBracket[winner] = tournament.WinnersBracket[winner] + 1

	tournament.MatchHistory = append(tournament.MatchHistory, result)
}

// func createBracketImage(tournament *DoubleElimTournament) image.Image {
// 	const imgWidth, imgHeight = 1000, 700
// 	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
// 	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

// 	// Draw match boxes
// 	red := color.RGBA{209, 123, 142, 255}
// 	blue := color.RGBA{143, 175, 204, 255}
// 	unplayed := color.RGBA{168, 168, 168, 255}

// 	type BoxInfo struct {
// 		x, y                             int
// 		matchLabel, alliance1, alliance2 string
// 		boxColor                         color.Color
// 	}
// 	var boxInfo []BoxInfo
// 	for i := 0; i < 7; i++ {
// 		matchNum := strconv.Itoa(i + 1)
// 		if i == 5 {
// 			matchNum = "FINAL"
// 		} else if i == 6 {
// 			matchNum = "FINAL #2"
// 		}

// 		if i < len(tournament.MatchHistory) {
// 			match := tournament.MatchHistory[i]
// 			if match.Winner == Red {
// 				boxInfo = append(boxInfo, BoxInfo{x: 50, y: 100, matchLabel: "M" + matchNum, alliance1: fmt.Sprintf("%d", match.RedAlliance.Captain), alliance2: fmt.Sprintf("%d", match.RedAlliance.FirstPick), boxColor: red})
// 			} else if match.Winner == Blue {
// 				boxInfo = append(boxInfo, BoxInfo{x: 50, y: 200, matchLabel: "M" + matchNum, alliance1: fmt.Sprintf("%d", match.BlueAlliance.Captain), alliance2: fmt.Sprintf("%d", match.BlueAlliance.FirstPick), boxColor: blue})
// 			}
// 		} else {
// 			boxInfo = append(boxInfo, BoxInfo{x: 50, y: 300, matchLabel: "M" + matchNum, alliance1: "", alliance2: "", boxColor: unplayed})
// 		}
// 	}
// 	for i := 0; i < len(boxInfo); i++ {
// 		drawMatchBox(img, boxInfo[i].x, boxInfo[i].y, boxInfo[i].matchLabel, boxInfo[i].alliance1, boxInfo[i].alliance2, boxInfo[i].boxColor)
// 	}

// 	// Draw connecting lines
// 	drawLine(img, 200, 125, 250, 175)
// 	drawLine(img, 200, 225, 250, 325)
// 	drawLine(img, 400, 175, 450, 250)
// 	drawLine(img, 400, 325, 450, 250)
// 	drawLine(img, 600, 250, 650, 250)
// 	drawLine(img, 800, 250, 850, 250)

// 	return img
// }
