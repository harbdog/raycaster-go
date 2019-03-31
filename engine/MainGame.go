package engine

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"raycaster-go/engine/raycaster"
	"runtime"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	// ebiten constants
	screenWidth  = 1024
	screenHeight = 700

	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texSize = 256
)

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
	//private GraphicsDeviceManager graphics;
	//private SpriteBatch spriteBatch;

	textures [5]*ebiten.Image

	//--test texture--//
	floor *ebiten.Image
	sky   *ebiten.Image

	//-test effect--//
	//Effect effect;

	//test sprite
	sprite *ebiten.Image

	//--array of levels, levels reffer to "floors" of the world--//
	levels []*Level
}

type Level struct {
	//--struct to represent rects and tints of a level--//
	Sv  []*image.Rectangle
	Cts []*image.Rectangle

	//--current slice tint (for lighting)--//
	St         []*color.Color
	CurrTexNum []int
}

func NewGame() *Game {
	// initialize Game object
	g := new(Game)

	g.width = screenWidth
	g.height = screenHeight

	g.slicer = raycaster.NewTextureHandler(texSize)

	return g
}

func (g *Game) Run() {
	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}

	if err := ebiten.Run(g.Update, g.width, g.height, 1, "Raycaster-Go"); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update(screen *ebiten.Image) error {
	g.view = screen

	// Perform logical updates
	mx, my := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		fmt.Printf("mouse left clicked: (%v, %v)\n", mx, my)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		mx, my := ebiten.CursorPosition()
		fmt.Printf("mouse right clicked: (%v, %v)\n", mx, my)
	}

	if ebiten.IsDrawingSkipped() {
		// When the game is running slowly, the rendering result
		// will not be adopted.
		return nil
	}

	// Render game to screen
	// TPS counter
	fps := fmt.Sprintf("TPS: %f/%v", ebiten.CurrentTPS(), ebiten.MaxTPS())
	ebitenutil.DebugPrint(g.view, fps)

	return nil
}
