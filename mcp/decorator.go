package mcp

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sedyh/ebitengine-mcp/internal/event"
)

const (
	timeout = 500 * time.Millisecond
	alive   = 5 * time.Second
)

type state int

const (
	start state = iota
	record
	exit
)

type decorator struct {
	game   ebiten.Game
	server *Server
	state  state
	next   *time.Ticker
	frames int
	pngs   []string
	err    error
}

func (g *decorator) Update() error {
	if g.state == exit {
		res := &event.RecordResponse{Images: g.pngs}
		res.SetError(g.err)
		g.server.RecordResponce(res)
		<-time.After(timeout)
		return ebiten.Termination
	}
	select {
	case req := <-g.server.RecordRequests():
		g.state = record
		g.frames = req.Frames
		g.next = time.NewTicker(req.Delay)
	default:
	}

	return g.game.Update()
}

func (g *decorator) Draw(screen *ebiten.Image) {
	g.game.Draw(screen)
	if g.state != record {
		return
	}
	select {
	case <-g.next.C:
		var buf bytes.Buffer
		if err := png.Encode(&buf, screen); err != nil {
			g.err = err
			g.next.Stop()
			g.state = exit
			return
		}
		g.pngs = append(g.pngs, base64.StdEncoding.EncodeToString(buf.Bytes()))
		g.frames--
		if g.frames == 0 {
			g.next.Stop()
			g.state = exit
		}
	default:
	}
}

func (g *decorator) Layout(w, h int) (int, int) {
	return g.game.Layout(w, h)
}

func Wrap(game ebiten.Game) ebiten.Game {
	server, err := NewServer()
	if err != nil {
		return NewInstantFailGame(err)
	}
	game = NewPostCloseGame(game, server.Close)

	return &decorator{
		game:   game,
		server: server,
	}
}
