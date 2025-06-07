package main

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sedyh/ebitengine-mcp/mcp"
	"golang.org/x/image/colornames"
)

func main() {
	log.SetFlags(log.Ltime)
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(mcp.Wrap(NewGame())); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	count float64
	size  float64
}

func NewGame() *Game {
	return &Game{
		count: -math.Pi / 1.2,
		size:  80.,
	}
}

func (g *Game) Update() error {
	log.Printf("fps: %.f\n", ebiten.ActualFPS())
	g.count += 0.1
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Gainsboro)
	cx, cy := screen.Bounds().Dx()/2, screen.Bounds().Dy()/2
	x, y := float32(math.Cos(g.count)*g.size*2), float32(math.Sin(g.count)*g.size*2)
	vector.DrawFilledCircle(screen, float32(cx)+x, float32(cy)+y, float32(g.size), colornames.Goldenrod, true)
}

func (g *Game) Layout(w, h int) (int, int) {
	return w, h
}

func (g *Game) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geo ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{GeoM: geo}
	op.ColorScale.Scale(0.25, 0.8, 0.5, 1)
	screen.DrawImage(offscreen, op)
}
