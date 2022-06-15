# raycaster-go

Golang raycasting engine using the [Ebitengine 2D Game Library](https://github.com/hajimehoshi/ebiten) to render a
3-Dimensional perspective in a 2-Dimensional view. It was originally converted from the C# based
[OwlRaycastEngine](https://github.com/Owlzy/OwlRaycastEngine), which in turn was created based on a
[raycasting tutorial](https://lodev.org/cgtutor/raycasting.html).

To see it in action visit the [YouTube playlist](https://www.youtube.com/playlist?list=PLOINtzQqJWIh8OQsvYAahr2yuAF5VLk38).

![Screenshot](https://raw.githubusercontent.com/harbdog/raycaster-go-demo/main/docs/images/screenshot.jpg)

## Demo

The [raycaster-go-demo](https://github.com/harbdog/raycaster-go-demo) project is available as an example of how to
use the raycaster-go engine as a module.

## Developer setup

To get started with your own project using the raycaster-go engine as a Go module:

1. Download, install, and setup Golang https://golang.org/dl/
2. Setup your project to use go modules: `go mod init github.com/yourname/yourgame` 
3. Download the raycaster-go module: `go get github.com/harbdog/raycaster-go`

**NOTE**: Depending on the OS, the Ebiten game library may have
[additional dependencies to install](https://ebiten.org/documents/install.html).

## Creating your own raycaster project

You will first want to become familiar with how to use the [Ebitengine 2D Game Library](https://ebiten.org/).
It has all of the APIs needed to render images on a 2D canvas, handle inputs from the player, and even play sounds.

*More to come* - There are some Golang interface functions that are required for your project to provide the necessary
information to render using the raycaster-go engine. Stay tuned... for now, see the
[raycaster-go-demo](https://github.com/harbdog/raycaster-go-demo) for how to setup a basic Ebitengine game loop
using the raycaster-go engine module.


## Limitations

* Raycasting is not raytracing.
* Raycasting draws 2D textures and sprites using a semi-3D technique, not using 3D models.
* The raycasting technique used in this project is more like early raycaster games such as Wolfenstein 3D,
  as opposed to later games such as Doom - it does not support stairs, sloped walls,
  or differing heights in elevation levels.
* Multiple elevation levels can be rendered, however player and sprite movement needs to be limited to the ground level.
* Only a single repeating floor texture can currently be set for the entire map.
* [Ceiling textures](https://lodev.org/cgtutor/raycasting2.html) are not yet implemented. Skybox texture
  is currently the only option, so going indoors from outdoors is not yet possible.
* [Thin walls](https://lodev.org/cgtutor/raycasting4.html#Thin), [doors]((https://lodev.org/cgtutor/raycasting4.html#Doors)),
  and [secret push walls](https://lodev.org/cgtutor/raycasting4.html#Secrets) are not currently implemented,
  feel free to figure them out and contribute as a Pull Request!
* [Translucent sprites](https://lodev.org/cgtutor/raycasting3.html#Translucent) are not currently implemented,
  feel free to contribute as a Pull Request!
