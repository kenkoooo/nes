package ui

import (
	"image"

	"../nes"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const padding = 0

type GameView struct {
	window  *glfw.Window
	audio   *Audio
	console *nes.Console
	title   string
	texture uint32
	record  bool
	frames  []image.Image
}

func NewGameView(window *glfw.Window, audio *Audio, console *nes.Console, title string) *GameView {
	texture := createTexture()
	return &GameView{window, audio, console, title, texture, false, nil}
}

func (view *GameView) Enter() {
	gl.ClearColor(0, 0, 0, 1)
	view.window.SetTitle(view.title)
	view.console.SetAudioChannel(view.audio.channel)
	view.console.SetAudioSampleRate(view.audio.sampleRate)
	view.window.SetKeyCallback(view.onKey)

	view.console.Reset()
}

func (view *GameView) Exit() {
	view.window.SetKeyCallback(nil)
	view.console.SetAudioChannel(nil)
	view.console.SetAudioSampleRate(0)
}

func (view *GameView) Update(t, dt float64) {
	if dt > 1 {
		dt = 0
	}
	window := view.window
	console := view.console

	updateControllers(window, console)
	console.StepSeconds(dt)
	gl.BindTexture(gl.TEXTURE_2D, view.texture)
	setTexture(console.Buffer())
	drawBuffer(view.window)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	if view.record {
		view.frames = append(view.frames, copyImage(console.Buffer()))
	}
}

func (view *GameView) onKey(window *glfw.Window,
	key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeySpace:
			screenshot(view.console.Buffer())
		case glfw.KeyR:
			view.console.Reset()
		case glfw.KeyTab:
			if view.record {
				view.record = false
				animation(view.frames)
				view.frames = nil
			} else {
				view.record = true
			}
		}
	}
}

func drawBuffer(window *glfw.Window) {
	w, h := window.GetFramebufferSize()
	s1 := float32(w) / 256
	s2 := float32(h) / 240
	f := float32(1 - padding)
	var x, y float32
	if s1 >= s2 {
		x = f * s2 / s1
		y = f
	} else {
		x = f
		y = f * s1 / s2
	}
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 1)
	gl.Vertex2f(-x, -y)
	gl.TexCoord2f(1, 1)
	gl.Vertex2f(x, -y)
	gl.TexCoord2f(1, 0)
	gl.Vertex2f(x, y)
	gl.TexCoord2f(0, 0)
	gl.Vertex2f(-x, y)
	gl.End()
}

func updateControllers(window *glfw.Window, console *nes.Console) {
	turbo := console.PPU.Frame%6 < 3
	k1 := readKeys(window, turbo)
	j1 := readJoystick(glfw.Joystick1, turbo)
	j2 := readJoystick(glfw.Joystick2, turbo)
	console.SetButtons1(combineButtons(k1, j1))
	console.SetButtons2(j2)
}
