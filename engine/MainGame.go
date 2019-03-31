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

	//--init texture slices--//
	g.slices = g.slicer.GetSlices()

	//--inits the levels--//
	g.levels = g.createLevels(4)

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

//returns an initialised Level struct
func (g *Game) createLevels(numLevels int) []*Level {
	var arr []*Level
	arr = make([]*Level, numLevels)

	for i := 0; i < numLevels; i++ {
		arr[i] = new(Level)
		arr[i].Sv = g.sliceView()
		arr[i].Cts = make([]*image.Rectangle, g.width)
		arr[i].St = make([]*color.Color, g.width)
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
