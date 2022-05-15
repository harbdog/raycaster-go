package engine

import (
	"image"
	"image/color"
)

func (g *Game) miniMap() *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, g.mapWidth, g.mapHeight))

	// wall/world positions
	for x, row := range g.worldMap {
		for y := range row {
			c := g.getMapColor(x, y)
			if c.A == 255 {
				c.A = 142
			}
			m.Set(x, y, c)
		}
	}

	// sprite positions
	sprites := g.getSprites()
	for sIndex, sprite := range sprites {
		// TODO: add map color variable to sprite object
		var spriteColor color.RGBA
		switch sIndex {
		case 0:
			spriteColor = color.RGBA{255, 200, 0, 142}
		default:
			spriteColor = color.RGBA{9, 70, 0, 142}
		}

		m.Set(int(sprite.Pos.X), int(sprite.Pos.Y), spriteColor)
	}

	// player position
	m.Set(int(g.player.Pos.X), int(g.player.Pos.Y), color.RGBA{255, 0, 0, 255})

	return m
}

func (g *Game) getMapColor(x, y int) color.RGBA {
	switch g.worldMap[x][y] {
	case 0:
		return color.RGBA{43, 30, 24, 255}
	case 1:
		return color.RGBA{100, 89, 73, 255}
	case 2:
		return color.RGBA{51, 32, 0, 196}
	case 3:
		return color.RGBA{56, 36, 0, 196}
	default:
		return color.RGBA{255, 194, 32, 255}
	}
}
