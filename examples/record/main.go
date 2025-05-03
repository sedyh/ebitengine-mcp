package main

import (
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"

	"github.com/sedyh/ebitengine-mcp/mcp"
)

func main() {
	ebiten.SetWindowSize(800, 600)
	if err := ebiten.RunGame(mcp.Wrap(&Game{count: -math.Pi / 1.2, lastUpdate: time.Now()})); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	count      float64
	lastUpdate time.Time
}

func (g *Game) Update() error {
	currentTime := time.Now()
	elapsed := currentTime.Sub(g.lastUpdate).Seconds()
	g.count += 1.0 * elapsed // Adjust animation speed as needed
	g.lastUpdate = currentTime
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Gainsboro)
	size := 150.
	cx, cy := screen.Bounds().Dx()/2, screen.Bounds().Dy()/2
	x, y := float32(math.Cos(g.count)*size*2), float32(math.Sin(g.count)*size*2)
	vector.DrawFilledCircle(screen, float32(cx)+x, float32(cy)+y, float32(size), colornames.Goldenrod, true)
}

func (g *Game) Layout(w, h int) (int, int) {
	return w, h
}
