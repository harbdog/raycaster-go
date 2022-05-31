package model

import (
	"image"
	"image/color"
	_ "image/png"
	"math"
	"sort"

	"raycaster-go/engine/geom"

	"github.com/hajimehoshi/ebiten/v2"
)

type Sprite struct {
	*Entity
	W, H           int
	Scale, vOffset float64
	AnimationRate  int
	animReversed   bool
	animCounter    int
	loopCounter    int
	columns, rows  int
	texNum, lenTex int
	texFacingMap   map[float64]int
	texFacingKeys  []float64
	texRects       []image.Rectangle
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
	uSize int, vOffset, collisionRadius float64,
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
	s.vOffset = vOffset

	s.texNum = 0
	s.lenTex = 1
	s.textures = make([]*ebiten.Image, s.lenTex)

	s.W, s.H = img.Size()
	s.texRects = []image.Rectangle{image.Rect(0, 0, s.W, s.H)}

	s.textures[0] = img

	return s
}

func NewSpriteFromSheet(
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA,
	columns, rows, spriteIndex int, uSize int, vOffset, collisionRadius float64,
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
	s.vOffset = vOffset

	s.texNum = spriteIndex
	s.columns, s.rows = columns, rows
	s.lenTex = columns * rows
	s.textures = make([]*ebiten.Image, s.lenTex)
	s.texRects = make([]image.Rectangle, s.lenTex)

	w, h := img.Size()

	// crop sheet by given number of columns and rows into a single dimension array
	s.W = w / columns
	s.H = h / rows

	for r := 0; r < rows; r++ {
		y := r * s.H
		for c := 0; c < columns; c++ {
			x := c * s.W
			cellRect := image.Rect(x, y, x+s.W-1, y+s.H-1)
			cellImg := img.SubImage(cellRect).(*ebiten.Image)

			index := c + r*columns
			s.textures[index] = cellImg
			s.texRects[index] = cellRect
		}
	}

	return s
}

func NewAnimatedSprite(
	x, y, scale float64, animationRate int, img *ebiten.Image, mapColor color.RGBA,
	columns, rows int, uSize int, vOffset, collisionRadius float64,
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
	s.vOffset = vOffset

	s.AnimationRate = animationRate
	s.animCounter = 0
	s.loopCounter = 0

	s.texNum = 0
	s.columns, s.rows = columns, rows
	s.lenTex = columns * rows
	s.textures = make([]*ebiten.Image, s.lenTex)
	s.texRects = make([]image.Rectangle, s.lenTex)

	w, h := img.Size()

	// crop sheet by given number of columns and rows into a single dimension array
	s.W = w / columns
	s.H = h / rows

	for r := 0; r < rows; r++ {
		y := r * s.H
		for c := 0; c < columns; c++ {
			x := c * s.W
			cellRect := image.Rect(x, y, x+s.W-1, y+s.H-1)
			cellImg := img.SubImage(cellRect).(*ebiten.Image)

			index := c + r*columns
			s.textures[index] = cellImg
			s.texRects[index] = cellRect
		}
	}

	return s
}

func (s *Sprite) SetTextureFacingMap(texFacingMap map[float64]int) {
	s.texFacingMap = texFacingMap

	// create pre-sorted list of keys used during facing determination
	s.texFacingKeys = make([]float64, len(texFacingMap))
	for k := range texFacingMap {
		s.texFacingKeys = append(s.texFacingKeys, k)
		if k == 0 {
			// duplicate entry at 0 to 2*Pi to match on higher angles
			texFacingMap[geom.Pi2] = texFacingMap[k]
			s.texFacingKeys = append(s.texFacingKeys, geom.Pi2)
		}
	}
	sort.Float64s(s.texFacingKeys)
}

func (s *Sprite) getTextureFacingKeyForAngle(facingAngle float64) float64 {
	var closestKeyAngle float64 = -1
	if s.texFacingMap == nil || len(s.texFacingMap) == 0 || s.texFacingKeys == nil || len(s.texFacingKeys) == 0 {
		return closestKeyAngle
	}

	closestKeyDiff := math.MaxFloat64
	for _, keyAngle := range s.texFacingKeys {
		keyDiff := math.Abs(float64(keyAngle) - facingAngle)
		if keyDiff < closestKeyDiff {
			closestKeyDiff = keyDiff
			closestKeyAngle = keyAngle
		}
	}

	return closestKeyAngle
}

func (s *Sprite) SetAnimationReversed(isReverse bool) {
	s.animReversed = isReverse
}

func (s *Sprite) SetAnimationFrame(texNum int) {
	s.texNum = texNum
}

func (s *Sprite) GetLoopCounter() int {
	return s.loopCounter
}

func (s *Sprite) GetVerticalOffset() float64 {
	return s.vOffset
}

func (s *Sprite) Update(camPos *geom.Vector2) {
	if s.AnimationRate <= 0 {
		return
	}

	if s.animCounter >= s.AnimationRate {
		minTexNum := 0
		maxTexNum := s.lenTex - 1

		if len(s.texFacingMap) > 1 && camPos != nil {
			// TODO: may want to be able to change facing even between animation frame changes

			// use facing from camera position to determine min/max texNum in texFacingMap
			// to update facing of sprite relative to camera and sprite angle
			texRow := 0

			// calculate angle from sprite relative to camera position by getting angle of line between them
			lineToCam := geom.Line{X1: s.Pos.X, Y1: s.Pos.Y, X2: camPos.X, Y2: camPos.Y}
			facingAngle := lineToCam.Angle() - s.Angle
			if facingAngle < 0 {
				// convert to positive angle needed to determine facing index to use
				facingAngle += geom.Pi2
			}
			facingKeyAngle := s.getTextureFacingKeyForAngle(facingAngle)
			if texFacingValue, ok := s.texFacingMap[facingKeyAngle]; ok {
				texRow = texFacingValue
			}

			minTexNum = texRow * s.columns
			maxTexNum = texRow*s.columns + s.columns - 1
		}

		s.animCounter = 0

		if s.animReversed {
			s.texNum -= 1
			if s.texNum > maxTexNum || s.texNum < minTexNum {
				s.texNum = maxTexNum
				s.loopCounter++
			}
		} else {
			s.texNum += 1
			if s.texNum > maxTexNum || s.texNum < minTexNum {
				s.texNum = minTexNum
				s.loopCounter++
			}
		}
	} else {
		s.animCounter++
	}
}

func (s *Sprite) GetTexture() *ebiten.Image {
	return s.textures[s.texNum]
}

func (s *Sprite) GetTextureRect() image.Rectangle {
	return s.texRects[s.texNum]
}
