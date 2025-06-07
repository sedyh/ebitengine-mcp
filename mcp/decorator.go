package mcp

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sedyh/ebitengine-mcp/internal/event"
)

func Wrap(game ebiten.Game) ebiten.Game {
	server, err := NewServer()
	if err != nil {
		return NewInstantFailGame(err)
	}

	dec := NewRecorderDecorator(game, server)
	dec = NewCloserDecorator(dec, server.Close)

	return NewOptionalGame(game, dec)
}

type closer struct {
	dec Decorator
	f   func()
}

func NewCloserDecorator(dec Decorator, f func()) Decorator {
	return &closer{dec: dec, f: f}
}

func (g *closer) Setup(op SetupOptions) {
	g.dec.Setup(op)
}

func (g *closer) Update() error {
	if err := g.dec.Update(); err != nil {
		g.f()
		return err
	}
	return nil
}

func (g *closer) Draw(screen *ebiten.Image) {
	g.dec.Draw(screen)
}

type recorder struct {
	game   ebiten.Game
	server *Server
	state  State
	next   *time.Ticker
	frames int
	pngs   []string
	err    error
	draw   bool
}

func NewRecorderDecorator(game ebiten.Game, server *Server) Decorator {
	return &recorder{game: game, server: server}
}

func (g *recorder) Setup(op SetupOptions) {
	g.draw = op.Draw
}

func (g *recorder) Update() error {
	if g.state == Exit {
		res := &event.RecordResponse{Images: g.pngs}
		res.SetError(g.err)
		g.server.RecordResponce(res)
		<-time.After(Timeout)
		return ebiten.Termination
	}
	select {
	case req := <-g.server.RecordRequests():
		g.state = Record
		g.frames = req.Frames
		g.next = time.NewTicker(req.Delay)
	default:
	}

	return g.game.Update()
}

func (g *recorder) Draw(screen *ebiten.Image) {
	if g.draw {
		g.game.Draw(screen)
	}
	if g.state != Record {
		return
	}
	select {
	case <-g.next.C:
		var buf bytes.Buffer
		if err := png.Encode(&buf, screen); err != nil {
			g.err = err
			g.next.Stop()
			g.state = Exit
			return
		}
		g.pngs = append(g.pngs, base64.StdEncoding.EncodeToString(buf.Bytes()))
		g.frames--
		if g.frames == 0 {
			g.next.Stop()
			g.state = Exit
		}
	default:
	}
}
