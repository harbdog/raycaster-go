package raycaster

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Draw the raycasted camera view to the screen.
func (c *Camera) Draw(screen *ebiten.Image) {
	screen.Clear()

	//--draw basic sky and floor--//
	texRect := image.Rect(0, 0, c.texSize, c.texSize)
	whiteRGBA := &color.RGBA{255, 255, 255, 255}

	floorRect := image.Rect(0, int(float64(c.h)*0.5)+c.pitch,
		c.w, c.h)
	drawTexture(screen, c.floor, &floorRect, &texRect, whiteRGBA)

	skyRect := image.Rect(0, 0, c.w, int(float64(c.h)*0.5)+c.pitch)
	drawTexture(screen, c.sky, &skyRect, &texRect, whiteRGBA)

	//--draw walls--//
	for x := 0; x < c.w; x++ {
		for i := cap(c.levels) - 1; i >= 0; i-- {
			drawTexture(screen, c.levels[i].CurrTex[x], c.levels[i].Sv[x], c.levels[i].Cts[x], c.levels[i].St[x])
		}
	}

	// draw textured floor
	floorImg := ebiten.NewImageFromImage(c.floorLvl.horBuffer)
	if floorImg == nil {
		log.Fatal("floorImg is nil")
	} else {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(floorImg, op)
	}

	// draw sprites
	for x := 0; x < c.w; x++ {
		for i := 0; i < cap(c.spriteLvls); i++ {
			spriteLvl := c.spriteLvls[i]
			if spriteLvl == nil {
				continue
			}

			texture := spriteLvl.CurrTex[x]
			if texture != nil {
				drawTexture(screen, texture, spriteLvl.Sv[x], spriteLvl.Cts[x], spriteLvl.St[x])
			}
		}
	}

	// FPS/TPS counter
	fps := fmt.Sprintf("FPS: %f\nTPS: %f/%v", ebiten.CurrentFPS(), ebiten.CurrentTPS(), ebiten.MaxTPS())
	ebitenutil.DebugPrint(screen, fps)
}

func drawTexture(screen *ebiten.Image, texture *ebiten.Image, destinationRectangle *image.Rectangle, sourceRectangle *image.Rectangle, color *color.RGBA) {
	if texture == nil || destinationRectangle == nil || sourceRectangle == nil {
		return
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

	screen.DrawImage(destTexture, op)
}
