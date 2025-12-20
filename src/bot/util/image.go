package util

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

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
