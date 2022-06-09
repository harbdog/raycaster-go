package engine

import (
	"image"
	"image/color"
	"sort"

	"github.com/harbdog/raycaster-go/engine/model"
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

	// sprite positions, sort by color to avoid random color getting chosen as last when using map keys
	sprites := make([]*model.Entity, 0, len(g.sprites))
	for s := range g.sprites {
		sprites = append(sprites, s.Entity)
	}
	sort.Slice(sprites, func(i, j int) bool {
		iComp := (sprites[i].MapColor.R + sprites[i].MapColor.G + sprites[i].MapColor.B)
		jComp := (sprites[j].MapColor.R + sprites[j].MapColor.G + sprites[j].MapColor.B)
		return iComp < jComp
	})

	for _, sprite := range sprites {
		if sprite.MapColor.A > 0 {

			m.Set(int(sprite.Pos.X), int(sprite.Pos.Y), sprite.MapColor)
		}
	}

	// projectile positions
	projectiles := make([]*model.Entity, 0, len(g.projectiles))
	for p := range g.projectiles {
		projectiles = append(projectiles, p.Entity)
	}
	sort.Slice(projectiles, func(i, j int) bool {
		iComp := (projectiles[i].MapColor.R + projectiles[i].MapColor.G + projectiles[i].MapColor.B)
		jComp := (projectiles[j].MapColor.R + projectiles[j].MapColor.G + projectiles[j].MapColor.B)
		return iComp < jComp
	})

	for _, projectile := range projectiles {
		if projectile.MapColor.A > 0 {

			m.Set(int(projectile.Pos.X), int(projectile.Pos.Y), projectile.MapColor)
		}
	}

	// player position
	m.Set(int(g.player.Pos.X), int(g.player.Pos.Y), g.player.MapColor)

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
