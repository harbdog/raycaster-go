package engine

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth * renderScale, screenHeight * renderScale
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	g.view = screen
	g.view.Clear()

	// Update camera (calculate raycast)
	g.camera.Update()

	//--draw basic sky and floor--//
	texRect := image.Rect(0, 0, texSize, texSize)
	whiteRGBA := &color.RGBA{255, 255, 255, 255}

	floorRect := image.Rect(0, int(float64(g.height)*0.5)+g.camera.GetPitch(),
		g.width, 2*int(float64(g.height)*0.5)-g.camera.GetPitch())
	g.spriteBatch.drawTexture(g.floor, &floorRect, &texRect, whiteRGBA)

	skyRect := image.Rect(0, 0, g.width, int(float64(g.height)*0.5)+g.camera.GetPitch())
	g.spriteBatch.drawTexture(g.sky, &skyRect, &texRect, whiteRGBA)

	//--draw walls--//
	for x := 0; x < g.width; x++ {
		for i := cap(g.levels) - 1; i >= 0; i-- {
			g.spriteBatch.drawTexture(g.levels[i].CurrTex[x], g.levels[i].Sv[x], g.levels[i].Cts[x], g.levels[i].St[x])
		}
	}

	// draw textured floor
	floorImg := ebiten.NewImageFromImage(g.floorLvl.HorBuffer)
	if floorImg == nil {
		log.Fatal("floorImg is nil")
	} else {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterLinear
		g.view.DrawImage(floorImg, op)
	}

	// draw sprites
	for x := 0; x < g.width; x++ {
		for i := 0; i < cap(g.spriteLvls); i++ {
			spriteLvl := g.spriteLvls[i]
			if spriteLvl == nil {
				continue
			}

			texture := spriteLvl.CurrTex[x]
			if texture != nil {
				g.spriteBatch.drawTexture(texture, spriteLvl.Sv[x], spriteLvl.Cts[x], spriteLvl.St[x])
			}
		}
	}

	// draw minimap
	mm := g.miniMap()
	mmImg := ebiten.NewImageFromImage(mm)
	if mmImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		op.GeoM.Scale(5.0, 5.0)
		op.GeoM.Translate(0, 50)
		view := g.view
		view.DrawImage(mmImg, op)
	}

	// draw crosshairs
	if g.crosshairs != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		op.GeoM.Scale(g.crosshairScale, g.crosshairScale)
		op.GeoM.Translate(
			float64(g.width)/2-float64(g.crosshairs.W)*g.crosshairScale/2,
			float64(g.height)/2-float64(g.crosshairs.H)*g.crosshairScale/2,
		)
		view := g.view
		view.DrawImage(g.crosshairs.GetTexture(), op)
	}

	if g.DebugOnce {
		// end DebugOnce after one loop
		g.DebugOnce = false
	}

	// draw for debugging
	if g.DebugX >= 0 && g.DebugY >= 0 {
		fX := float64(g.DebugX)
		fY := float64(g.DebugY)
		// draw a red translucent dot at the debug point
		ebitenutil.DrawLine(g.view, fX-0.5, fY-0.5, fX+0.5, fY+0.5, color.RGBA{255, 0, 0, 150})

		// draw two red vertical lines focusing on point
		ebitenutil.DrawLine(g.view, fX-0.5, fY+5, fX+0.5, fY+25, color.RGBA{255, 0, 0, 150})
		ebitenutil.DrawLine(g.view, fX-0.5, fY-25, fX+0.5, fY-5, color.RGBA{255, 0, 0, 150})
	}

	// FPS/TPS counter
	fps := fmt.Sprintf("FPS: %f\nTPS: %f/%v", ebiten.CurrentFPS(), ebiten.CurrentTPS(), ebiten.MaxTPS())
	ebitenutil.DebugPrint(g.view, fps)
}

func (s *SpriteBatch) drawTexture(texture *ebiten.Image, destinationRectangle *image.Rectangle, sourceRectangle *image.Rectangle, color *color.RGBA) {
	if texture == nil || destinationRectangle == nil || sourceRectangle == nil {
		return
	}

	if sourceRectangle.Min.X == 0 {
		// fixes subImage from clipping at edges of textures which can cause gaps
		sourceRectangle.Min.X++
		sourceRectangle.Max.X++
	}

	// if destinationRectangle is not the same size as sourceRectangle, scale to fit
	var scaleX, scaleY float64 = 1.0, 1.0
	if !destinationRectangle.Eq(*sourceRectangle) {
		sSize := sourceRectangle.Size()
		dSize := destinationRectangle.Size()

		scaleX = float64(dSize.X) / float64(sSize.X)
		scaleY = float64(dSize.Y) / float64(sSize.Y)
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear

	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(float64(destinationRectangle.Min.X), float64(destinationRectangle.Min.Y))

	destTexture := texture.SubImage(*sourceRectangle).(*ebiten.Image)

	if color != nil {
		// color channel modulation/tinting
		op.ColorM.Scale(float64(color.R)/255, float64(color.G)/255, float64(color.B)/255, float64(color.A)/255)
	}

	if s.g.DebugX > destinationRectangle.Min.X && s.g.DebugX <= destinationRectangle.Max.X &&
		s.g.DebugY > destinationRectangle.Min.Y && s.g.DebugY <= destinationRectangle.Max.Y {

		for texNum, tex := range s.g.tex.Textures {
			if tex == texture {
				s.g.DebugPrintfOnce("[draw@%v,%v]: %v | %v < %v * %v,%v\n", s.g.DebugX, s.g.DebugY, destinationRectangle, texNum, sourceRectangle, scaleX, scaleY)
				return
			}
		}
	}

	view := s.g.view
	view.DrawImage(destTexture, op)
}
