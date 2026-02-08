package bot

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/shuban-789/bjorn/src/bot/interactions"
	"github.com/shuban-789/bjorn/src/bot/search"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type BracketTracker struct {
	Year              string
	EventCode         string
	EventName         string
	Matches           map[int]*BracketMatch
	Alliances         map[int]*Alliance
	Champion          *Alliance
	ProcessedMatchIDs map[int]bool
	mu                sync.Mutex
}

type Alliance struct {
	Seed      int
	Captain   int
	FirstPick int
}

type BracketMatch struct {
	MatchID      int
	Series       int
	RedAlliance  *Alliance
	BlueAlliance *Alliance
	RedScore     int
	BlueScore    int
	Winner       *Alliance
	Loser        *Alliance
	Played       bool
}

var (
	bracketBgColor     = color.RGBA{32, 34, 37, 255}
	bracketLineColor   = color.RGBA{185, 187, 190, 255}
	bracketBoxColor    = color.RGBA{64, 68, 75, 255}
	bracketBoxBorder   = color.RGBA{114, 137, 218, 255}
	bracketRedColor    = color.RGBA{255, 82, 82, 255}
	bracketBlueColor   = color.RGBA{85, 170, 255, 255}
	bracketWinnerColor = color.RGBA{67, 181, 129, 255}
	bracketTextColor   = color.RGBA{255, 255, 255, 255}
	bracketGrayColor   = color.RGBA{185, 187, 190, 255}
	bracketTitleColor  = color.RGBA{255, 215, 0, 255}
	bracketLoserColor  = color.RGBA{240, 71, 71, 255}
)

var (
	bracketTrackers = make(map[string]*BracketTracker)
	bracketMu       sync.RWMutex
)

func GetOrCreateBracketTracker(year, eventCode string) *BracketTracker {
	key := year + "-" + eventCode
	bracketMu.Lock()
	defer bracketMu.Unlock()

	if tracker, exists := bracketTrackers[key]; exists {
		return tracker
	}

	// todo: hardcoding san diego for now, not ideal though
	var eventName string = search.GetEventNameFromCode("USCASD", eventCode)
	if eventName == "" {
		eventName = eventCode
	}

	tracker := &BracketTracker{
		Year:              year,
		EventCode:         eventCode,
		EventName:         eventName,
		Matches:           make(map[int]*BracketMatch),
		Alliances:         make(map[int]*Alliance),
		ProcessedMatchIDs: make(map[int]bool),
	}
	bracketTrackers[key] = tracker
	return tracker
}

func (bt *BracketTracker) UpdateBracketWithMatch(matchID, series int, redTeams, blueTeams []TeamDTO, redScore, blueScore int) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	if bt.ProcessedMatchIDs[matchID] {
		return
	}

	match, exists := bt.Matches[series]
	if !exists {
		match = &BracketMatch{Series: series}
		bt.Matches[series] = match
	}

	match.RedAlliance = bt.findOrCreateAlliance(redTeams)
	match.BlueAlliance = bt.findOrCreateAlliance(blueTeams)
	match.MatchID = matchID
	match.RedScore = redScore
	match.BlueScore = blueScore
	match.Played = true

	if redScore > blueScore {
		match.Winner = match.RedAlliance
		match.Loser = match.BlueAlliance
	} else {
		match.Winner = match.BlueAlliance
		match.Loser = match.RedAlliance
	}

	bt.ProcessedMatchIDs[matchID] = true

	// todo: make this properly handle double elim finals, currently just assumes series 6 and 7 are finals and that winner of either is champion
	if series == 9 {
		bt.Champion = match.Winner
	} else if series == 10 {
		bt.Champion = match.Winner
	}
}

func (bt *BracketTracker) findOrCreateAlliance(teams []TeamDTO) *Alliance {
	var captain, firstPick int
	for _, t := range teams {
		if t.AllianceRole == "Captain" {
			captain = t.TeamNumber
		} else if t.AllianceRole == "FirstPick" {
			firstPick = t.TeamNumber
		}
	}

	if alliance, exists := bt.Alliances[captain]; exists {
		return alliance
	}
	seed := len(bt.Alliances) + 1
	alliance := &Alliance{
		Seed:      seed,
		Captain:   captain,
		FirstPick: firstPick,
	}
	bt.Alliances[captain] = alliance
	return alliance
}

func (bt *BracketTracker) GenerateBracketImage() (*bytes.Buffer, error) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	width := 720
	height := 340

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.Draw(img, img.Bounds(), &image.Uniform{bracketBgColor}, image.Point{}, draw.Src)

	title := bt.EventName + " PLAYOFFS BRACKET"
	drawTextOnBracket(img, 250, 18, title, bracketTitleColor)

	drawTextOnBracket(img, 50, 45, "UPPER BRACKET", bracketGrayColor)
	drawTextOnBracket(img, 50, 235, "LOWER BRACKET", bracketGrayColor)
	drawTextOnBracket(img, 510, 125, "FINALS", bracketGrayColor)
	positions := map[int]struct{ x, y int }{
		1: {50, 55},
		2: {50, 135},
		3: {280, 95},
		4: {50, 245},
		5: {280, 245},
		6: {510, 135},
		7: {510, 215},
	}

	labels := map[int]string{
		1: "Upper R1-A",
		2: "Upper R1-B",
		3: "Upper Final",
		4: "Lower R1",
		5: "Lower Final",
		6: "Finals 1",
		7: "Finals 2",
	}

	boxWidth := 160
	boxHeight := 55

	for series := 1; series <= 7; series++ {
		pos := positions[series]
		match := bt.Matches[series]
		drawMatchBox(img, pos.x, pos.y, boxWidth, boxHeight, labels[series], match)
	}

	if bt.Champion != nil {
		cx, cy := 510, 55
		rect := image.Rect(cx, cy, cx+boxWidth, cy+boxHeight)
		draw.Draw(img, rect, &image.Uniform{color.RGBA{40, 100, 60, 255}}, image.Point{}, draw.Src)
		drawBoxBorder(img, cx, cy, boxWidth, boxHeight, bracketTitleColor, 3)
		drawTextOnBracket(img, cx+30, cy+20, "CHAMPION", bracketTitleColor)
		drawTextOnBracket(img, cx+20, cy+38, formatAlliance(bt.Champion), bracketWinnerColor)
	}

	drawThickLineOnBracket(img, 210, 82, 280, 122, bracketLineColor)
	drawThickLineOnBracket(img, 210, 162, 280, 122, bracketLineColor)
	drawThickLineOnBracket(img, 210, 272, 280, 272, bracketLineColor)
	drawThickLineOnBracket(img, 360, 150, 360, 245, bracketLineColor)
	drawThickLineOnBracket(img, 440, 122, 510, 162, bracketLineColor)
	drawThickLineOnBracket(img, 440, 272, 485, 272, bracketLineColor)
	drawThickLineOnBracket(img, 485, 272, 485, 190, bracketLineColor)
	drawThickLineOnBracket(img, 485, 190, 510, 190, bracketLineColor)
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func formatAlliance(a *Alliance) string {
	if a == nil {
		return "TBD"
	}
	return fmt.Sprintf("%d & %d", a.Captain, a.FirstPick)
}

func drawMatchBox(img *image.RGBA, x, y, w, h int, label string, match *BracketMatch) {
	rect := image.Rect(x, y, x+w, y+h)
	draw.Draw(img, rect, &image.Uniform{bracketBoxColor}, image.Point{}, draw.Src)
	drawBoxBorder(img, x, y, w, h, bracketBoxBorder, 2)
	drawTextOnBracket(img, x+5, y+14, label, bracketGrayColor)

	if match == nil || !match.Played {
		redText := "TBD"
		blueText := "TBD"
		if match != nil {
			if match.RedAlliance != nil {
				redText = formatAlliance(match.RedAlliance)
			}
			if match.BlueAlliance != nil {
				blueText = formatAlliance(match.BlueAlliance)
			}
		}
		drawTextOnBracket(img, x+5, y+30, redText, bracketRedColor)
		drawTextOnBracket(img, x+5, y+46, blueText, bracketBlueColor)
		return
	}

	redText := fmt.Sprintf("%s (%d)", formatAlliance(match.RedAlliance), match.RedScore)
	blueText := fmt.Sprintf("%s (%d)", formatAlliance(match.BlueAlliance), match.BlueScore)

	redColor := bracketLoserColor
	blueColor := bracketLoserColor
	if match.Winner == match.RedAlliance {
		redText = "W " + redText
		redColor = bracketWinnerColor
		blueText = "L " + blueText
	} else {
		blueText = "W " + blueText
		blueColor = bracketWinnerColor
		redText = "L " + redText
	}

	drawTextOnBracket(img, x+5, y+30, redText, redColor)
	drawTextOnBracket(img, x+5, y+46, blueText, blueColor)
}

func drawBoxBorder(img *image.RGBA, x, y, w, h int, col color.Color, thickness int) {
	for i := 0; i < thickness; i++ {
		drawLineOnBracket(img, x+i, y+i, x+w-i, y+i, col)
		drawLineOnBracket(img, x+i, y+h-i, x+w-i, y+h-i, col)
		drawLineOnBracket(img, x+i, y+i, x+i, y+h-i, col)
		drawLineOnBracket(img, x+w-i, y+i, x+w-i, y+h-i, col)
	}
}

func drawTextOnBracket(img *image.RGBA, x, y int, text string, col color.Color) {
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
	face := basicfont.Face7x13
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  point,
	}
	d.DrawString(text)
}

func drawLineOnBracket(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	dx := absB(x2 - x1)
	dy := absB(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x1, y1, col)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func drawThickLineOnBracket(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	drawLineOnBracket(img, x1, y1, x2, y2, col)
	drawLineOnBracket(img, x1, y1+1, x2, y2+1, col)
	drawLineOnBracket(img, x1, y1-1, x2, y2-1, col)
}

func absB(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func handleBracketCommand(session *discordgo.Session, i *discordgo.InteractionCreate, year, eventCode string) {
	channelID := i.ChannelID

	tracker := GetOrCreateBracketTracker(year, eventCode)

	imgBuf, err := tracker.GenerateBracketImage()
	if err != nil {
		interactions.SendMessage(session, i, channelID, fmt.Sprintf("Failed to generate bracket: %v", err))
		return
	}

	playedCount := 0
	tracker.mu.Lock()
	for _, m := range tracker.Matches {
		if m.Played {
			playedCount++
		}
	}
	champion := tracker.Champion
	tracker.mu.Unlock()

	var description string
	if playedCount == 0 {
		description = "*No playoff matches have been recorded yet.*\n\n" +
			"Start event tracking with `/match eventstart` or `/match track` " +
			"and the bracket will update automatically as playoff matches are played."
	} else {
		description = fmt.Sprintf("**%d playoff match(es) completed**\n\n", playedCount)
		if champion != nil {
			description += fmt.Sprintf("🏆 **Champion: %s**", formatAlliance(champion))
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🏆 %s Playoffs Bracket", eventCode),
		Description: description,
		Color:       0x7289DA,
		Image: &discordgo.MessageEmbedImage{
			URL: "attachment://bracket.png",
		},
	}

	_, err = session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
		Files:  []*discordgo.File{{Name: "bracket.png", Reader: imgBuf}},
	})
	if err != nil {
		_, _ = session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Embed: embed,
			Files: []*discordgo.File{{Name: "bracket.png", Reader: imgBuf}},
		})
	}
}
