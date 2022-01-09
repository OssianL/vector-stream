package main

import (
	"fmt"
	"time"

	"github.com/goxjs/gl"
	"github.com/goxjs/glfw"
	"github.com/shibukawa/nanovgo"
)

var blowup bool
var premult bool

func key(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	} else if key == glfw.KeySpace && action == glfw.Press {
		blowup = !blowup
	} else if key == glfw.KeyP && action == glfw.Press {
		premult = !premult
	}
}

func main() {

	err := glfw.Init(gl.ContextWatcher)
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// demo MSAA
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "NanoVGo", nil, nil)
	if err != nil {
		panic(err)
	}
	window.SetKeyCallback(key)
	window.MakeContextCurrent()

	ctx, err := nanovgo.NewContext(0 /*nanovgo.AntiAlias | nanovgo.StencilStrokes | nanovgo.Debug*/)
	defer ctx.Delete()

	if err != nil {
		panic(err)
	}

	glfw.SwapInterval(0)

	client := NewClient(ctx)
	server := NewServer()

	funDefBytes := server.Init()
	funDefBytecode := NewBytecodeFromBytes(funDefBytes)
	fmt.Println(funDefBytecode)
	client.Update(funDefBytecode)
	fmt.Println("init done")

	for !window.ShouldClose() {
		time.Sleep(time.Millisecond * 5)

		fbWidth, fbHeight := window.GetFramebufferSize()
		winWidth, winHeight := window.GetSize()
		// mx, my := window.GetCursorPos()

		pixelRatio := float32(fbWidth) / float32(winWidth)
		gl.Viewport(0, 0, fbWidth, fbHeight)
		if premult {
			gl.ClearColor(0, 0, 0, 0)
		} else {
			gl.ClearColor(0.3, 0.3, 0.32, 1.0)
		}
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.Enable(gl.CULL_FACE)
		gl.Disable(gl.DEPTH_TEST)

		ctx.BeginFrame(winWidth, winHeight, pixelRatio)

		fmt.Println("new update frame")

		bytes := server.Update()
		bytecode := NewBytecodeFromBytes(bytes)
		client.Update(bytecode)
		client.Render()

		ctx.EndFrame()

		gl.Enable(gl.DEPTH_TEST)
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
