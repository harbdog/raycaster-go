package engine

import (
	"fmt"
	"github.com/harbdog/raycaster-go/engine/raycaster"
	"image"
	"image/color"
	"log"
	"math"
	"path/filepath"
	"runtime"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	// ebiten constants
	screenWidth  = 1024
	screenHeight = 700
	screenScale  = 1.0

	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texSize = 256
)

// Game - This is the main type for your game.
type Game struct {
	//--create slicer and declare slices--//
	slicer *raycaster.TextureHandler
	slices []*image.Rectangle

	//--viewport and width / height--//
	view   *ebiten.Image
	width  int
	height int

	//--define camera--//
	camera *raycaster.Camera

	//--graphics manager and sprite batch--//
	spriteBatch *SpriteBatch

	textures [5]*ebiten.Image

	//--test texture--//
	floor *ebiten.Image
	sky   *ebiten.Image

	//-test effect--//
	//Effect effect;

	//test sprite
	sprite *ebiten.Image

	//--array of levels, levels reffer to "floors" of the world--//
	levels []*raycaster.Level

	// for debugging
	DebugX    int
	DebugY    int
	DebugOnce bool
}

// SpriteBatch - converted C# method Graphics.SpriteBatch
// Enables a group of sprites to be drawn using the same settings.
type SpriteBatch struct {
	g *Game
}

// NewGame - Allows the game to perform any initialization it needs to before starting to run.
// This is where it can query for any required services and load any non-graphic
// related content.  Calling base.Initialize will enumerate through any components
// and initialize them as well.
func NewGame() *Game {
	// initialize Game object
	g := new(Game)

	// use scale to keep the desired window width and height
	g.width = int(math.Floor(float64(screenWidth) / screenScale))
	g.height = int(math.Floor(float64(screenHeight) / screenScale))

	g.slicer = raycaster.NewTextureHandler(texSize)

	//--init texture slices--//
	g.slices = g.slicer.GetSlices()

	//--inits the levels--//
	g.levels = g.createLevels(4)

	//--init camera--//
	g.camera = raycaster.NewCamera(g.width, g.height, texSize, g.slices, g.levels)

	// for debugging
	g.DebugX = -1
	g.DebugY = -1

	return g
}

// loadContent will be called once per game and is the place to load
// all of your content.
func (g *Game) loadContent() {
	// Create a new SpriteBatch, which can be used to draw textures.
	g.spriteBatch = &SpriteBatch{g: g}

	// TODO: use this.Content to load your game content here
	g.textures[0], _, _ = getTextureFromFile("stone.png")
	g.textures[1], _, _ = getTextureFromFile("left_bot_house.png")
	g.textures[2], _, _ = getTextureFromFile("right_bot_house.png")
	g.textures[3], _, _ = getTextureFromFile("left_top_house.png")
	g.textures[4], _, _ = getTextureFromFile("right_top_house.png")

	g.floor, _, _ = getTextureFromFile("floor.png")
	g.sky, _, _ = getTextureFromFile("sky.png")
}

func getTextureFromFile(texFile string) (*ebiten.Image, image.Image, error) {
	resourcePath := filepath.Join("engine", "content", "textures")
	eImg, iImg, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, texFile), ebiten.FilterNearest)
	if err != nil {
		log.Fatal(err)
	}
	return eImg, iImg, err
}

// Run is the Ebiten Run loop caller
func (g *Game) Run() {
	// load content once when first run
	g.loadContent()

	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}

	if err := ebiten.Run(g.Update, g.width, g.height, screenScale, "Raycaster-Go"); err != nil {
		log.Fatal(err)
	}
}

// Update - Allows the game to run logic such as updating the world,
// checking for collisions, gathering input, and playing audio.
func (g *Game) Update(screen *ebiten.Image) error {
	g.view = screen

	// Perform logical updates
	g.camera.Update()

	// TODO: Add your update logic here
	g.handleInput()

	if ebiten.IsDrawingSkipped() {
		// When the game is running slowly, the rendering result
		// will not be adopted.
		return nil
	}

	// Render game to screen
	g.draw()

	// TPS counter
	fps := fmt.Sprintf("TPS: %f/%v", ebiten.CurrentTPS(), ebiten.MaxTPS())
	ebitenutil.DebugPrint(g.view, fps)

	return nil
}

func (g *Game) handleInput() {
	mx, my := ebiten.CursorPosition()

	forward := false
	backward := false
	rotLeft := false
	rotRight := false

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		fmt.Printf("mouse left clicked: (%v, %v)\n", mx, my)

		// using left click for debugging graphical issues
		if g.DebugX == -1 && g.DebugY == -1 {
			// only allow setting once between clears to debounce
			g.DebugX = mx
			g.DebugY = my
			g.DebugOnce = true
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		fmt.Printf("mouse right clicked: (%v, %v)\n", mx, my)

		// using right click to clear the debugging flag
		g.DebugX = -1
		g.DebugY = -1
		g.DebugOnce = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		rotLeft = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		rotRight = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		forward = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		backward = true
	}

	if forward {
		g.camera.Move(0.06)
	} else if backward {
		g.camera.Move(-0.06)
	}

	if rotLeft {
		g.camera.Rotate(0.03)
	} else if rotRight {
		g.camera.Rotate(-0.03)
	}
}

func (g *Game) draw() {
	g.view.Clear()

	//--draw sky and floor--//
	texRect := image.Rect(0, 0, texSize, texSize)
	whiteRGBA := &color.RGBA{255, 255, 255, 255}

	// spriteBatch.Draw(floor,
	//    new Rectangle(0, (int)(height * 0.5f), width, (int)(height * 0.5f)),
	//    new Rectangle(0, 0, texSize, texSize),
	//    Color.White);
	floorRect := image.Rect(0, int(float64(g.height)*0.5), g.width, 2*int(float64(g.height)*0.5))
	g.spriteBatch.draw(g.floor, &floorRect, &texRect, whiteRGBA)

	// spriteBatch.Draw(sky,
	//    new Rectangle(0, 0, width, (int)(height * 0.5f)),
	//    new Rectangle(0, 0, texSize, texSize),
	//    Color.White);
	skyRect := image.Rect(0, 0, g.width, int(float64(g.height)*0.5))
	g.spriteBatch.draw(g.sky, &skyRect, &texRect, whiteRGBA)

	//--draw walls--//
	for x := 0; x < g.width; x++ {
		for i := cap(g.levels) - 1; i >= 0; i-- {
			g.spriteBatch.draw(g.textures[g.levels[i].CurrTexNum[x]], g.levels[i].Sv[x], g.levels[i].Cts[x], g.levels[i].St[x])
		}
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
}

//returns an initialised Level struct
func (g *Game) createLevels(numLevels int) []*raycaster.Level {
	var arr []*raycaster.Level
	arr = make([]*raycaster.Level, numLevels)

	for i := 0; i < numLevels; i++ {
		arr[i] = new(raycaster.Level)
		arr[i].Sv = g.sliceView()
		arr[i].Cts = make([]*image.Rectangle, g.width)
		arr[i].St = make([]*color.RGBA, g.width)
		arr[i].CurrTexNum = make([]int, g.width)

		for j := 0; j < cap(arr[i].CurrTexNum); j++ {
			arr[i].CurrTexNum[j] = 1
		}
	}

	return arr
}

// Creates rectangle slices for each x in width.
func (g *Game) sliceView() []*image.Rectangle {
	var arr []*image.Rectangle
	arr = make([]*image.Rectangle, g.width)

	for x := 0; x < g.width; x++ {
		thisRect := image.Rect(x, 0, x+1, g.height)
		arr[x] = &thisRect
	}

	return arr
}

func (s *SpriteBatch) draw(texture *ebiten.Image, destinationRectangle *image.Rectangle, sourceRectangle *image.Rectangle, color *color.RGBA) {
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear

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

	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(float64(destinationRectangle.Min.X), float64(destinationRectangle.Min.Y))

	var destTexture *ebiten.Image
	destTexture = texture.SubImage(*sourceRectangle).(*ebiten.Image)

	// color channel modulation/tinting
	op.ColorM.Scale(float64(color.R)/255, float64(color.G)/255, float64(color.B)/255, float64(color.A)/255)

	if s.g.DebugX > destinationRectangle.Min.X && s.g.DebugX <= destinationRectangle.Max.X {
		for texNum, tex := range s.g.textures {
			if tex == texture {
				s.g.DebugPrintfOnce("[debug@%v,%v]: %v | %v < %v * %v,%v\n", s.g.DebugX, s.g.DebugY, destinationRectangle, texNum, sourceRectangle, scaleX, scaleY)
				return
			}
		}
	}

	view := s.g.view
	view.DrawImage(destTexture, op)
}

// DebugPrintfOnce prints info to screen only one time until g.DebugFlag cleared again
func (g *Game) DebugPrintfOnce(format string, a ...interface{}) {
	if g.DebugOnce {
		fmt.Printf(format, a...)
	}
}
