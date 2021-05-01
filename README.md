# raycaster-go
Golang raycaster engine based on the [raycaster example](https://github.com/faiface/pixel-examples/blob/master/community/raycaster/raycaster.go) for the [Pixel](https://github.com/faiface/pixel) 2D Game Library.

![Screenshot](screenshot.png?raw=true)

## Developer setup
To install and run from source the following are required:
1. Download, install, and setup Golang https://golang.org/dl/
2. Windows requires gcc: [mingw-w64](https://mingw-w64.org) toolchain (you can download
   [this file](https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-posix/seh/x86_64-8.1.0-release-posix-seh-rt_v6-rev0.7z)
   in particular).
2. Use the `go get` command to download the Pixel 2D Game Library and its pre-requisites: 
    * `[GOPATH]$  go get -u github.com/faiface/pixel`
    * `[GOPATH]$  go get -u github.com/faiface/glhf`
    * `[GOPATH]$  go get -u github.com/faiface/mainthread`
    * `[GOPATH]$  go get -u github.com/go-gl/mathgl/mgl32`
    * `[GOPATH]$  go get -u github.com/go-gl/glfw/v3.3/glfw`
3. Clone/download this project to your `GOPATH/src` folder, for example: `GOPATH/src/raycaster-go`
4. Now you can use the `go run` command to run `raycast.go`:
    * `[GOPATH/src/raycaster-go]$ go run raycast.go`

## Controls
* Move and rotate using WASD or Arrow Keys
