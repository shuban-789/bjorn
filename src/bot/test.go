//go:build ignore
// +build ignore

package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type AllianceColor int

const (
	Red AllianceColor = iota
	Blue
	Neither
)

type TwoTeamAlliance struct {
	Captain   int
	FirstPick int
	Color     AllianceColor
}

type MatchResult struct {
	RedAlliance  TwoTeamAlliance
	BlueAlliance TwoTeamAlliance
	RedScore     int
	BlueScore    int
	Winner       AllianceColor
	MatchNumber  string
}

type DoubleElimTournament struct {
	WinnersBracket map[TwoTeamAlliance]int
	LosersBracket  map[TwoTeamAlliance]int
	Eliminated     map[TwoTeamAlliance]bool
	MatchHistory   []MatchResult
}

func drawText(img *image.RGBA, x, y int, label string) {
	col := color.Black
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
	face := basicfont.Face7x13
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  point,
	}
	d.DrawString(label)
}

func drawMatchBox(img *image.RGBA, x, y int, matchLabel string, alliance1, alliance2 string, boxColor color.Color) {
	rect := image.Rect(x, y, x+150, y+50)
	draw.Draw(img, rect, &image.Uniform{boxColor}, image.Point{}, draw.Src)
	drawText(img, x+10, y+15, matchLabel)
	drawText(img, x+10, y+30, alliance1)
	drawText(img, x+10, y+45, alliance2)
}

// Function to draw a line between two points
func drawLine(img *image.RGBA, x1, y1, x2, y2 int) {
	// Simple line drawing (Bresenham's line algorithm)
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
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
		img.Set(x1, y1, color.RGBA{0, 0, 0, 255})
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

// Helper function to calculate absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func createBracketImage(tournament *DoubleElimTournament) image.Image {
	const imgWidth, imgHeight = 1000, 700
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw match boxes
	red := color.RGBA{209, 123, 142, 255}
	blue := color.RGBA{143, 175, 204, 255}
	unplayed := color.RGBA{168, 168, 168, 255}
	drawMatchBox(img, 50, 100, "M1", "A1", "A4", red)
	drawMatchBox(img, 50, 200, "M2", "A2", "A3", blue)
	drawMatchBox(img, 250, 150, "M4", "", "", unplayed)
	drawMatchBox(img, 250, 300, "M3", "", "", unplayed)
	drawMatchBox(img, 450, 225, "M5", "", "", unplayed)
	drawMatchBox(img, 650, 225, "M6", "", "", unplayed)
	drawMatchBox(img, 850, 225, "M7 (optional)", "", "", unplayed)

	// Draw connecting lines
	drawLine(img, 200, 125, 250, 175)
	drawLine(img, 200, 225, 250, 325)
	drawLine(img, 400, 175, 450, 250)
	drawLine(img, 400, 325, 450, 250)
	drawLine(img, 600, 250, 650, 250)
	drawLine(img, 800, 250, 850, 250)

	return img
}

func main() {
	tournament := &DoubleElimTournament{}
	img := createBracketImage(tournament)
	file, err := os.Create("bracket.png")
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()
	png.Encode(file, img)
}
