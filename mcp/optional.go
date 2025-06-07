package mcp

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

func NewOptionalGame(
	src ebiten.Game,
	dec Decorator,
) ebiten.Game {
	if v, ok := src.(ebiten.FinalScreenDrawer); ok {
		dec.Setup(SetupOptions{Draw: false})
		return &final{
			game:  src,
			dec:   dec,
			final: v,
		}
	}
	dec.Setup(SetupOptions{Draw: true})
	return &plain{
		game: src,
		dec:  dec,
	}
}

type plain struct {
	game ebiten.Game
	dec  Decorator
}

func (g *plain) Update() error {
	if err := g.game.Update(); err != nil {
		return err
	}
	if err := g.dec.Update(); err != nil {
		return err
	}
	return nil
}

func (g *plain) Draw(screen *ebiten.Image) {
	g.game.Draw(screen)
	g.dec.Draw(screen)
}

func (g *plain) Layout(w, h int) (int, int) {
	return g.game.Layout(w, h)
}

type final struct {
	game      ebiten.Game
	dec       Decorator
	final     ebiten.FinalScreenDrawer
	offscreen *ebiten.Image
}

func (g *final) Update() error {
	if err := g.game.Update(); err != nil {
		return err
	}
	if err := g.dec.Update(); err != nil {
		return err
	}
	return nil
}

func (g *final) Draw(screen *ebiten.Image) {
	g.game.Draw(screen)
}

func (g *final) Layout(w, h int) (int, int) {
	return g.game.Layout(w, h)
}

func (g *final) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geo ebiten.GeoM) {
	ow, oh := screen.Bounds().Dx(), screen.Bounds().Dy()
	if g.offscreen != nil && g.changed(ow, oh) {
		g.offscreen.Deallocate()
		g.offscreen = nil
	}
	if g.offscreen == nil {
		g.offscreen = ebiten.NewImageWithOptions(
			image.Rect(0, 0, ow, oh),
			&ebiten.NewImageOptions{Unmanaged: true},
		)
	}

	g.offscreen.Clear()
	g.final.DrawFinalScreen(g.offscreen, offscreen, geo)
	screen.DrawImage(g.offscreen, nil)
	g.dec.Draw(g.offscreen)
}

func (g *final) changed(w, h int) bool {
	return g.offscreen.Bounds().Dx() != w || g.offscreen.Bounds().Dy() != h
}

type fail struct {
	err error
}

func NewInstantFailGame(err error) ebiten.Game {
	return &fail{err: err}
}

func (g *fail) Update() error {
	return g.err
}

func (g *fail) Draw(screen *ebiten.Image) {}

func (g *fail) Layout(w, h int) (int, int) {
	return w, h
}
