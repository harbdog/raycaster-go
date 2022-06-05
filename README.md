# raycaster-go
Golang conversion of raycaster engine [OwlRaycastEngine](https://github.com/Owlzy/OwlRaycastEngine)
using the [Ebiten 2D Game Library](https://github.com/hajimehoshi/ebiten).
To see it in action visit the [YouTube playlist](https://www.youtube.com/playlist?list=PLOINtzQqJWIh8OQsvYAahr2yuAF5VLk38).

![Screenshot](screenshot.jpg?raw=true)

## Developer setup
To install and run from source the following are required:
1. Download, install, and setup Golang https://golang.org/dl/
2. Clone/download this project locally.
3. From the project folder use the following command to download the Go module dependencies of this project:
    * `[path/to/raycaster-go]$ go mod download`
4. The Ebiten game library may have [additional dependencies to install](https://ebiten.org/documents/install.html),
   depending on the OS.
5. Now you can use the `go run` command to run `raycast.go`:
    * `[path/to/raycaster-go]$ go run raycast.go`

## Controls
* Press `Escape` key to toggle mouse cursor capture (or hold `CTRL`)
* Move the mouse to rotate and pitch view
* Move and strafe using `WASD` or `Arrow Keys`
* Click left mouse button to fire current weapon
* Use mouse wheel or press `1` or `2` to select a weapon
* Press `H` to holster/put away current weapon
* Hold `Shift` key to move faster
* Hold `ALT` key to enter mouse move mode (vertical mouse moves position instead of pitch)
* Hold `C` key for crouch position
* Hold `Z` key for prone position
* Hold `Spacebar` for jump position
