package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

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

func matchcmd(channelID string, args []string, session *discordgo.Session, guildID string, authorID string) {
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

		getMatch(channelID, year, eventCode, matchNumber, nil, session)
	case "eventstart":
		if len(args) < 3 {
			session.ChannelMessageSend(channelID, "Usage: `>>match eventstart <year> <eventCode>`")
			return
		}

		hasPerms, err := isAdmin(session, guildID, authorID)
		if err != nil {
			session.ChannelMessageSend(channelID, "Unable to check permissions of user.")
			return
		}

		if hasPerms {
			session.ChannelMessageSend(channelID, "You do not have permission to run this command.")
			return
		}

		year := args[1]
		eventCode := args[2]
		eventStart(channelID, guildID, year, eventCode, session)
	default:
		session.ChannelMessageSend(channelID, "Unknown subcommand. Available subcommands: `info`, `eventstart`")
	}
}

func eventStart(channelID, guildID, year, eventCode string, session *discordgo.Session) {
	eventDetails, err := fetchEventDetails(year, eventCode)
	if err != nil {
		session.ChannelMessageSend(channelID, err.Error())
		return
	}

	location, err := time.LoadLocation(eventDetails.TimeZone)
	if err != nil {
		session.ChannelMessageSend(channelID, fmt.Sprintf("Error loading event timezone: %v", err))
		return
	}

	today := time.Now().In(location)
	startTime, endTime, err := getEventStartEndTime(eventDetails, today, location)
	if err != nil {
		session.ChannelMessageSend(channelID, err.Error())
		return
	}

	if endTime.Before(today) {
		session.ChannelMessageSend(channelID, "This event has already ended!")
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
	session.ChannelMessageSend(channelID, fmt.Sprintf("Started tracking matches for event %s in %s...", eventCode, year))

	if startTime.Before(today) {
		session.ChannelMessageSend(channelID, "This event has already started! Will not create event.")
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
	session.ChannelMessageSend(channelID, fmt.Sprintf("Created event: %s", event.ID))
}

func getMatch(ChannelID string, year string, eventCode string, matchNumber string, event *EventTracked, session *discordgo.Session) {
	url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s/matches", year, eventCode)
	resp, err := http.Get(url)
	if err != nil {
		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to fetch match data: %v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
		Teams           []TeamDTO `json:"teams"`
		TournamentLevel string    `json:"tournamentLevel"`
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
		RedAlliance     TwoTeamAlliance
		BlueAlliance    TwoTeamAlliance
		TournamentLevel string
	}
	redTeams := []TeamDTO{}
	blueTeams := []TeamDTO{}
	for _, match := range matches {
		if fmt.Sprintf("%d", match.ID) == matchNumber {
			selectedMatch.Scores.Red = match.Scores.Red.TotalPoints
			selectedMatch.Scores.Blue = match.Scores.Blue.TotalPoints
			for _, team := range match.Teams {
				if team.AllianceColor == "Red" {
					redTeams = append(redTeams, team)
				} else if team.AllianceColor == "Blue" {
					blueTeams = append(blueTeams, team)
				} else {
					// TODO: I'm going to crash out if this happens
				}
				selectedMatch.TournamentLevel = match.TournamentLevel
			}
			break
		}
	}
	selectedMatch.RedAlliance = getAllianceFromTeams(redTeams)
	selectedMatch.BlueAlliance = getAllianceFromTeams(blueTeams)

	winner := "Red Alliance"
	winnerSkib := Red
	if selectedMatch.Scores.Blue > selectedMatch.Scores.Red {
		winnerSkib = Blue
		winner = "Blue Alliance"
	} else if selectedMatch.Scores.Blue == selectedMatch.Scores.Red {
		winnerSkib = Neither
		winner = "Draw"
	}

	color := 0xE02C44
	if winnerSkib == Blue {
		color = 0x58ACEC
	} else if winnerSkib == Neither {
		color = 0xE8E4EC
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s Match %s Results", eventCode, matchNumber),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Tournament Level",
				Value:  selectedMatch.TournamentLevel,
				Inline: false,
			},
			{
				Name:   "Red Alliance Teams  🔴   \u200B",
				Value:  fmt.Sprintf("%v", "Captain: "+fmt.Sprintf("%d", selectedMatch.RedAlliance.Captain)+"\nFirst Pick: "+fmt.Sprintf("%d", selectedMatch.RedAlliance.FirstPick)),
				Inline: true,
			},
			{
				Name:   "Blue Alliance Teams  🔵   \u200B",
				Value:  fmt.Sprintf("%v", "Captain: "+fmt.Sprintf("%d", selectedMatch.BlueAlliance.Captain)+"\nFirst Pick: "+fmt.Sprintf("%d", selectedMatch.BlueAlliance.FirstPick)),
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
	// 		session.ChannelMessageSend(ChannelID, fmt.Sprintf("Failed to encode image: %v", err))
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
}

func eventUpdate(apiPollTime time.Duration, session *discordgo.Session) {
	for i := 0; i < len(eventsBeingTracked); i++ {
		event := &eventsBeingTracked[i]

		fmt.Printf("\033[33m[INFO]\033[0m Checking event: Year=%s, EventCode=%s, LastUpdateTime=%v\n",
			event.Year, event.EventCode, event.LastUpdateTime)

		if time.Since(event.LastUpdateTime) < apiPollTime {
			fmt.Printf("\033[33m[INFO]\033[0m Skipping update for event %s/%s (last updated %v)\n",
				event.Year, event.EventCode, event.LastUpdateTime)
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
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Event %s/%s has started!", event.Year, event.EventCode))
		} else if !eventDetails.Ongoing && event.Ongoing {
			// remove the event once it's done
			eventsBeingTracked = append(eventsBeingTracked[:i], eventsBeingTracked[i+1:]...)
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("The %s has ended!", eventDetails.Name))
			return
		}

		url := fmt.Sprintf("https://api.ftcscout.org/rest/v1/events/%s/%s/matches", event.Year, event.EventCode)
		fmt.Printf("\033[33m[INFO]\033[0m Fetching matches from: %s\n", url)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("\033[31m[FAIL]\033[0m Failed to fetch match data for event %s/%s: %v\n",
				event.Year, event.EventCode, err)
			session.ChannelMessageSend(event.UpdateChannelId, fmt.Sprintf("Failed to fetch match data: %v", err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			fmt.Printf("\033[33m[INFO]\033[0m Event %s/%s not found!\n", event.Year, event.EventCode)
			session.ChannelMessageSend(event.UpdateChannelId, "That event does not exist!")
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("\033[31m[FAIL]\033[0m Failed to read response for event %s/%s: %v\n",
				event.Year, event.EventCode, err)
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

		fmt.Printf("\033[33m[INFO]\033[0m Matches fetched for event %s/%s: %d matches found\n",
			event.Year, event.EventCode, len(matches))

		var newMatches []Match
		for _, match := range matches {
			fmt.Printf("\033[33m[INFO]\033[0m Checking match: ID=%d, hasBeenPlayed=%v\n",
				match.ID, match.GetHasBeenPlayed())

			if match.ID > event.LastProcessedMatchId && match.GetHasBeenPlayed() {
				newMatches = append(newMatches, match)
				fmt.Printf("\033[33m[INFO]\033[0m New match to process: ID=%d\n", match.ID)
			}
		}

		if len(newMatches) > 0 {
			for _, match := range newMatches {
				fmt.Printf("\033[33m[INFO]\033[0m Processing match: ID=%d\n", match.ID)
				getMatch(event.UpdateChannelId, event.Year, event.EventCode, fmt.Sprintf("%d", match.ID), event, session)
			}
			event.LastProcessedMatchId = newMatches[len(newMatches)-1].ID
			fmt.Printf("\033[33m[INFO]\033[0m Updated LastProcessedMatchId for event %s/%s: %d\n",
				event.Year, event.EventCode, event.LastProcessedMatchId)
		} else {
			fmt.Println("\033[33m[INFO]\033[0m No new played matches found in this interval.")
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
		return EventDetails{}, fmt.Errorf("Failed to fetch match data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return EventDetails{}, fmt.Errorf("That event does not exist!")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return EventDetails{}, fmt.Errorf("Failed to read response: %v", err)
	}

	var eventDetails EventDetails
	err = json.Unmarshal(body, &eventDetails)
	if err != nil {
		return EventDetails{}, fmt.Errorf("Failed to parse event details: %v", err)
	}

	return eventDetails, nil
}

func getEventStartEndTime(eventDetails EventDetails, today time.Time, location *time.Location) (time.Time, time.Time, error) {
	layout := "2006-01-02"
	startTime, err := time.Parse(layout, eventDetails.Start)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Failed to parse event start time: %v", err)
	}

	endTime, err := time.Parse(layout, eventDetails.End)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Failed to parse event end time: %v", err)
	}

	startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 8, 0, 0, 0, location)
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 5, 0, 0, 0, location)

	// If we start after 8, set it to be scheduled in the future so it doesn't error out
	if startTime.Year() == today.Year() && startTime.YearDay() == today.YearDay() {
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
