package main

import (
	"image/color"

	"github.com/fogleman/gg"
)

func DrawStrokeText(imgGG *gg.Context, text string, x, y float64, textColor, strokeColor color.RGBA, strokeWidth float64) {
	imgGG.SetColor(strokeColor)
	for dy := -strokeWidth; dy <= strokeWidth; dy++ {
		for dx := -strokeWidth; dx <= strokeWidth; dx++ {
			if dx*dx+dy*dy >= strokeWidth*strokeWidth {
				// give it rounded corners
				continue
			}

			x := x + float64(dx)
			y := y + float64(dy)

			imgGG.DrawStringAnchored(text, x, y, 0.5, 0.5)
		}
	}

	imgGG.SetColor(textColor)
	imgGG.DrawStringAnchored(text, x, y, 0.5, 0.5)
}
