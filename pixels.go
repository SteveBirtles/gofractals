package main

import (
	_ "image/png"
	"log"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"time"
	"math/rand"
)

var (
	texture   uint32
	frameLength float64
	frames            = 0
	second            = time.Tick(time.Second)
	exit	bool
)


func main() {

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	window, err := glfw.CreateWindow(1920, 1080, "Pixels", glfw.GetPrimaryMonitor(), nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)

	window.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		switch {
		case key == glfw.KeyEscape && action == glfw.Press:
			exit = true
		}
	})


	if err := gl.Init(); err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	for !window.ShouldClose() && !exit {

		frameStart := time.Now()

		drawScene()
		window.SwapBuffers()
		glfw.PollEvents()

		frames++
		select {
		case <-second:
			frames = 0
		default:
		}

		frameLength = time.Since(frameStart).Seconds()
	}

	window.SetShouldClose(true)

}

func drawScene() {

	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0.0, 1920.0, 1080.0, 0.0, -1, 1)

	r, g, b := float32(1.0), float32(0.0), float32(1.0)

	gl.Begin(gl.POINTS)

	for x := int32(0); x < 1920; x++ {
		for y := int32(0); y < 1080; y++ {

			if x % 100 == 0 {
				r = 1.0
			}
			if x % 200 == 0 {
				r = 0.5
			}

			if y % 100 == 0 {
				b = 1.0
			}
			if y % 200 == 0 {
			 	b = 0.5
			}

			gl.Color3f(r, g, b)
			gl.Vertex2i(x, y)

		}
	}

	gl.End()

}
