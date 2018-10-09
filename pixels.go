package main

import (
	"fmt"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	_ "image/png"
	"log"
	"time"
	"runtime"
)

type pixel struct {
	r, g, b float32
}

const (
	WIDTH     = 1280
	HEIGHT    = 1024
	INFINITY  = 1e+50
	THREADS   = 3
	STARTSIZE = 64
	BATCH = WIDTH*16
)

var (
	texture        uint32
	frameLength    float64
	exit           bool
	screen         [WIDTH][HEIGHT]pixel
	processed      [WIDTH][HEIGHT]bool
	completed      chan int
	threadFinished [THREADS]bool
	xCentre            = WIDTH / 2
	yCentre            = HEIGHT / 2
	fracX                = -0.5//-1.4
	fracY                = 0.0//0.00056
	scale                = 0.005//0.0000001
	BLACK                = pixel{0.0, 0.0, 0.0}
	WHITE                = pixel{1.0, 1.0, 1.0}
	iterations      = 100//10000
)

func init() {
	runtime.LockOSThread()
}

func mandelbrot(x, y int) pixel {

	c := complex(float64(x-xCentre)*scale+fracX, float64(y-yCentre)*scale+fracY)

	z := complex(0, 0)

	var i int
	for i = 0; i < iterations; i++ {
		z = z*z + c
		if imag(z) > INFINITY || real(z) > INFINITY {
			break
		}
	}

	if i == iterations {
		return BLACK
	} else {

		value := float32(6*i) / float32(iterations)

		if value < 1 {
			return pixel{value, 0, 1}
		} else if value < 2 {
			value -= 1
			return pixel{1, value, 1 - value}
		} else if value < 3 {
			value -= 2
			return pixel{1 - value, 1, 0}
		} else if value < 4 {
			value -= 3
			return pixel{0, 1, value}
		} else if value < 5 {
			value -= 4
			return pixel{0, 1 - value, 1}
		} else {
			value -= 5
			return pixel{0, 0, 1 - value}
		}

	}

}

func render(core int) {

	fmt.Println("Thread", core, "started...")

	pixelSize := STARTSIZE
	renderX := 0
	renderY := core * pixelSize

	pixelCount := 0

renderLoop:
	for pixelCount < BATCH {

		if renderY < HEIGHT {

			if int(renderY/STARTSIZE)%THREADS == core {

				if !processed[renderX][renderY] {

					m := mandelbrot(renderX, renderY)
					pixelCount++

					for u := 0; u < pixelSize; u++ {
						for v := 0; v < pixelSize; v++ {
							if renderX+u < WIDTH && renderY+v < HEIGHT {
								screen[renderX+u][renderY+v] = m
							}
						}
					}
					processed[renderX][renderY] = true
				}

			}

			renderX += pixelSize

			if renderX >= WIDTH {
				renderX = 0
				renderY += pixelSize
			}

		} else {
			if pixelSize == 1 {
				fmt.Println("Thread", core, "is finished!!!")
				threadFinished[core] = true
				break renderLoop
			} else {
				renderX = 0
				renderY = core * pixelSize
				pixelSize = pixelSize / 2
				continue renderLoop
			}

		}

	}

	fmt.Println("Sending complete signal for thread", core)
	completed <- core


}

func drawScene() {

	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0.0, WIDTH, HEIGHT, 0.0, -1, 1)

	gl.MatrixMode(gl.MODELVIEW)

	gl.Begin(gl.POINTS)

	for t := 0; t < THREADS; t++ {
		for x := int32(0); x < WIDTH; x++ {
			for y := int32(0); y < HEIGHT; y++ {

				if screen[x][y] == BLACK {
					continue
				}
				gl.Color3f(screen[x][y].r, screen[x][y].g, screen[x][y].b)
				gl.Vertex2i(x, y)
			}
		}
	}


	gl.End()

}

func main() {

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	window, err := glfw.CreateWindow(WIDTH, HEIGHT, "Pixels", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	window.SetPos(0, 0)

	window.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		switch {
		case key == glfw.KeyEscape && action == glfw.Press:
			exit = true
		}
	})

	if err := gl.Init(); err != nil {
		panic(err)
	}

	completed = make(chan int, 1)

	for t := 0; t < THREADS; t++ {
		go render(t)
	}

	var done [THREADS]bool

	renderStart := time.Now()

	for !window.ShouldClose() && !exit {

		select {
		case c := <-completed:

			done[c] = true
			fmt.Println("Thread", c, "done.")
			doneCount := 0

			for t := 0; t < THREADS; t++ {
				if done[t] {
					doneCount++
				}
			}

			if doneCount == THREADS {

				fmt.Println("All threads done, rendering scene")
				drawScene()
				window.SwapBuffers()

				allFinished := true

				for t := 0; t < THREADS; t++ {
					done[c] = false
					if !threadFinished[t] {
						allFinished = false
						fmt.Println("Starting thread", t)
						go render(t)
					}
				}

				if allFinished {
					renderEnd := time.Since(renderStart).Seconds()
					fmt.Println("Render finished in", renderEnd, "seconds.")
				}

			}

		default:
		}

		glfw.PollEvents()

	}

	window.SetShouldClose(true)

}
