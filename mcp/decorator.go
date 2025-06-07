package mcp

import (
	"bytes"
	"encoding/base64"
	"flag"
	"image/png"
	"time"

	"github.com/sedyh/ebitengine-mcp/internal/cli"
	"github.com/sedyh/ebitengine-mcp/internal/srv"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	name    = "ebitengine-mcp-recorder-server"
	version = "1.0.0"
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
	req    srv.Events
	res    srv.Events
	state  state
	next   *time.Ticker
	start  time.Time
	frames int
	pngs   []string
	err    error
}

func (g *decorator) Update() error {
	// if time.Since(g.start) > alive && g.state == start {
	// 	return ebiten.Termination
	// }
	if g.state == exit {
		g.res[srv.Record] <- srv.RecordRes{
			Images: g.pngs,
			Err:    g.err,
		}
		<-time.After(timeout)
		return ebiten.Termination
	}
	select {
	case e := <-g.req[srv.Record]:
		switch e := e.(type) {
		case srv.RecordReq:
			g.state = record
			g.frames = e.Frames
			g.next = time.NewTicker(e.Delay)
		}
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
	d := &decorator{
		game:  game,
		req:   srv.NewEvents(srv.Record),
		res:   srv.NewEvents(srv.Record),
		start: time.Now(),
	}

	var enabled bool
	flag.BoolVar(&enabled, cli.FlagID, false, "enable mcp")
	flag.Parse()
	if enabled {
		go srv.Run(d.req, d.res)
	}

	return d
}
