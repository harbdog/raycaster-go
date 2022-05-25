package model

import (
	"image"
	"image/color"
	_ "image/png"

	"raycaster-go/engine/geom"

	"github.com/hajimehoshi/ebiten/v2"
)

type Sprite struct {
	*Entity
	W, H           int
	Scale          float64
	Anchor         SpriteAnchor
	AnimationRate  int
	animCounter    int
	loopCounter    int
	texNum, lenTex int
	textures       []*ebiten.Image
}

type SpriteAnchor int

const (
	BottomCenter SpriteAnchor = iota
	Center
	TopCenter
)

func NewSprite(
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA,
	uSize int, anchor SpriteAnchor, collisionRadius float64,
) *Sprite {
	s := &Sprite{
		Entity: &Entity{
			Pos:             &geom.Vector2{X: x, Y: y},
			PosZ:            0.5,
			Angle:           0,
			Velocity:        0,
			CollisionRadius: collisionRadius,
			MapColor:        mapColor,
		},
	}
	s.Scale = scale
	s.Anchor = anchor

	s.texNum = 0
	s.lenTex = 1
	s.textures = make([]*ebiten.Image, s.lenTex)

	// scale image if indicated
	if scale != 1.0 {
		w, h := img.Size()
		w = int(float64(w) * scale)
		h = int(float64(h) * scale)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)

		scaleImg := ebiten.NewImage(w, h)
		scaleImg.DrawImage(img, op)
		img = scaleImg
	}

	s.W, s.H = img.Size()
	if s.W != uSize || s.H != uSize {
		//translate image using anchor if not same size as 1u cell (texSize)
		op := &ebiten.DrawImageOptions{}
		translateX, translateY := getAnchorTranslate(anchor, s.W, s.H, uSize)
		op.GeoM.Translate(translateX, translateY)

		translateImg := ebiten.NewImage(uSize, uSize)
		translateImg.DrawImage(img, op)
		img = translateImg
	}

	s.textures[0] = img

	return s
}

func NewSpriteFromSheet(
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA,
	columns, rows, spriteIndex int, uSize int, anchor SpriteAnchor, collisionRadius float64,
) *Sprite {
	s := &Sprite{
		Entity: &Entity{
			Pos:             &geom.Vector2{X: x, Y: y},
			PosZ:            0.5,
			Angle:           0,
			Velocity:        0,
			CollisionRadius: collisionRadius,
			MapColor:        mapColor,
		},
	}
	s.Scale = scale
	s.Anchor = anchor

	s.texNum = spriteIndex
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
				// translate image using anchor if not same size as 1u cell (texSize)
				op2 := &ebiten.DrawImageOptions{}
				translateX, translateY := getAnchorTranslate(anchor, s.W, s.H, uSize)
				op2.GeoM.Translate(translateX, translateY)

				translateImg := ebiten.NewImage(uSize, uSize)
				translateImg.DrawImage(cellTarget, op2)
				cellTarget = translateImg
			}

			s.textures[c+r*columns] = cellTarget
		}
	}

	return s
}

func NewAnimatedSprite(
	x, y, scale float64, animationRate int, img *ebiten.Image, mapColor color.RGBA,
	columns, rows int, uSize int, anchor SpriteAnchor, collisionRadius float64,
) *Sprite {
	s := &Sprite{
		Entity: &Entity{
			Pos:             &geom.Vector2{X: x, Y: y},
			PosZ:            0.5,
			Angle:           0,
			Velocity:        0,
			CollisionRadius: collisionRadius,
			MapColor:        mapColor,
		},
	}
	s.Scale = scale
	s.Anchor = anchor

	s.AnimationRate = animationRate
	s.animCounter = 0
	s.loopCounter = 0

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
				// translate image using anchor if not same size as 1u cell (texSize)
				op2 := &ebiten.DrawImageOptions{}
				translateX, translateY := getAnchorTranslate(anchor, s.W, s.H, uSize)
				op2.GeoM.Translate(translateX, translateY)

				translateImg := ebiten.NewImage(uSize, uSize)
				translateImg.DrawImage(cellTarget, op2)
				cellTarget = translateImg
			}

			s.textures[c+r*columns] = cellTarget
		}
	}

	return s
}

func (s *Sprite) SetAnimationFrame(texNum int) {
	s.texNum = texNum
}

func (s *Sprite) GetLoopCounter() int {
	return s.loopCounter
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
			s.loopCounter++
		}
	} else {
		s.animCounter++
	}
}

func (s *Sprite) GetTexture() *ebiten.Image {
	return s.textures[s.texNum]
}

func getAnchorTranslate(anchor SpriteAnchor, spriteWidth, spriteHeight, unitSize int) (float64, float64) {

	switch anchor {
	case BottomCenter:
		return float64(unitSize/2 - spriteWidth/2), float64(unitSize - spriteHeight)
	case Center:
		return float64(unitSize/2 - spriteWidth/2), float64(unitSize-spriteHeight) / 2
	case TopCenter:
		return float64(unitSize/2 - spriteWidth/2), 0
	}

	return 0, 0
}
