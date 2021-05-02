package raycaster

import (
	"image"
	"time"

	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

type Sprite struct {
	X, Y           float64
	texNum, lenTex int
	textures       []*ebiten.Image
}

func NewSprite(x, y float64, img *ebiten.Image) *Sprite {
	s := &Sprite{}
	s.X, s.Y = x, y
	s.texNum = 0
	s.lenTex = 1
	s.textures = make([]*ebiten.Image, s.lenTex)
	s.textures[0] = img

	return s
}

func NewSpriteFromSheet(x, y float64, img *ebiten.Image, columns, rows int) *Sprite {
	s := &Sprite{}
	s.X, s.Y = x, y
	s.texNum = 0
	s.lenTex = columns * rows
	s.textures = make([]*ebiten.Image, s.lenTex)

	// crop sheet by given number of columns and rows into a single dimension array
	w, h := img.Size()
	wCell := w / columns
	hCell := h / rows

	op := &ebiten.DrawImageOptions{}

	for r := 0; r < rows; r++ {
		y := r * hCell
		for c := 0; c < columns; c++ {
			x := c * wCell
			cellRect := image.Rect(x, y, x+wCell-1, y+hCell-1)
			cellImg := img.SubImage(cellRect).(*ebiten.Image)

			cellTarget := ebiten.NewImage(wCell, hCell)
			cellTarget.DrawImage(cellImg, op)

			s.textures[c+r*c] = cellTarget
		}
	}

	// TESTING ANIMATION
	ticker := time.NewTicker(100 * time.Millisecond)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				s.X -= 0.1
				s.nextTexture()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return s
}

func (s *Sprite) nextTexture() {
	s.texNum += 1
	if s.texNum >= s.lenTex {
		s.texNum = 0
	}
}

func (s *Sprite) GetTexture() *ebiten.Image {
	return s.textures[s.texNum]
}
