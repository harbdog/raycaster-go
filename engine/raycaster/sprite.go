package raycaster

import (
	"image"
	_ "image/png"

	"raycaster-go/engine/geom"
	"raycaster-go/engine/model"

	"github.com/hajimehoshi/ebiten/v2"
)

type Sprite struct {
	*model.Entity
	Vx, Vy         float64
	W, H           int
	Scale          float64
	AnimationRate  int
	animCounter    int
	texNum, lenTex int
	textures       []*ebiten.Image
}

func NewSprite(x, y float64, img *ebiten.Image, uSize int, collisionRadius float64) *Sprite {
	s := &Sprite{
		Entity: &model.Entity{
			Pos:             &geom.Vector2{X: x, Y: y},
			Angle:           0,
			CollisionRadius: collisionRadius,
		},
	}
	s.Scale = 1.0
	s.texNum = 0
	s.lenTex = 1
	s.textures = make([]*ebiten.Image, s.lenTex)

	s.W, s.H = img.Size()
	if s.W != uSize || s.H != uSize {
		// translate image to center/bottom if not same size as 1u cell (texSize)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(uSize/2-s.W/2), float64(uSize-s.H))

		translateImg := ebiten.NewImage(uSize, uSize)
		translateImg.DrawImage(img, op)
		img = translateImg
	}

	s.textures[0] = img

	return s
}

func NewAnimatedSprite(
	x, y, scale float64, animationRate int, img *ebiten.Image, columns, rows int,
	uSize int, collisionRadius float64,
) *Sprite {
	s := &Sprite{
		Entity: &model.Entity{
			Pos:             &geom.Vector2{X: x, Y: y},
			Angle:           0,
			CollisionRadius: collisionRadius,
		},
	}
	s.Scale = scale
	s.AnimationRate = animationRate
	s.animCounter = 0

	s.texNum = 0
	s.lenTex = columns * rows
	s.textures = make([]*ebiten.Image, s.lenTex)

	w, h := img.Size()

	// scale image if indicated
	if scale != 1.0 {
		w = int(float64(w) * scale)
		h = int(float64(h) * scale)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)

		scaleImg := ebiten.NewImage(w, h)
		scaleImg.DrawImage(img, op)
		img = scaleImg
	}

	// crop sheet by given number of columns and rows into a single dimension array
	s.W = w / columns
	s.H = h / rows

	op := &ebiten.DrawImageOptions{}

	for r := 0; r < rows; r++ {
		y := r * s.H
		for c := 0; c < columns; c++ {
			x := c * s.W
			cellRect := image.Rect(x, y, x+s.W-1, y+s.H-1)
			cellImg := img.SubImage(cellRect).(*ebiten.Image)

			cellTarget := ebiten.NewImage(s.W, s.H)
			cellTarget.DrawImage(cellImg, op)

			if w != uSize || h != uSize {
				// translate image to center/bottom if not same size as 1u cell (texSize)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(uSize/2-s.W/2), float64(uSize-s.H))

				translateImg := ebiten.NewImage(uSize, uSize)
				translateImg.DrawImage(cellTarget, op)
				cellTarget = translateImg
			}

			s.textures[c+r*c] = cellTarget
		}
	}

	return s
}

func (s *Sprite) Update() {
	if s.AnimationRate <= 0 {
		return
	}

	if s.animCounter >= s.AnimationRate {
		s.animCounter = 0
		s.texNum += 1
		if s.texNum >= s.lenTex {
			s.texNum = 0
		}
	} else {
		s.animCounter++
	}
}

func (s *Sprite) GetTexture() *ebiten.Image {
	return s.textures[s.texNum]
}
