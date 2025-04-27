package main

import (
	"bytes"
	"flag"
	"image/png"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	command = "record"
	delay   = 300 * time.Millisecond
	frames  = 1
)

type McpResponse struct {
	Contents [][]byte `json:"contents,omitempty"`
	Error    string   `json:"error,omitempty"`
}

type Decorator struct {
	game      ebiten.Game
	count     int
	active    bool
	last      time.Time
	done      bool
	mcpResult McpResponse
}

func (g *Decorator) Update() error {
	if g.done {
		return ebiten.Termination
	}
	return g.game.Update()
}

func (g *Decorator) Draw(screen *ebiten.Image) {
	g.game.Draw(screen)

	if g.active && !g.done {
		now := time.Now()
		if now.Sub(g.last) >= delay {
			var buf bytes.Buffer
			if err := png.Encode(&buf, screen); err != nil {
				// Записываем первую ошибку и устанавливаем done
				if g.mcpResult.Error == "" {
					g.mcpResult.Error = err.Error()
				}
				g.done = true
				return
			}

			// Добавляем кадр в содержимое
			g.mcpResult.Contents = append(g.mcpResult.Contents, buf.Bytes())
			g.count++
			g.last = now

			// Проверяем, достигли ли мы заданного количества кадров
			if g.count >= frames {
				g.done = true
			}
		}
	}
}

func (g *Decorator) Layout(w, h int) (int, int) {
	return g.game.Layout(w, h)
}

func Wrap(game ebiten.Game) ebiten.Game {
	d := &Decorator{
		game: game,
		mcpResult: McpResponse{
			Contents: make([][]byte, 0, frames),
		},
	}
	flag.BoolVar(&d.active, command, false, "enable recording with automatic exit")
	flag.Parse()
	return d
}
