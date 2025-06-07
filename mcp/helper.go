package mcp

import (
	"github.com/hajimehoshi/ebiten/v2"
)

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

type closer struct {
	game ebiten.Game
	f    func()
}

func NewPostCloseGame(game ebiten.Game, f func()) ebiten.Game {
	return &closer{game, f}
}

func (g *closer) Update() error {
	if err := g.game.Update(); err != nil {
		g.f()
		return err
	}
	return nil
}

func (g *closer) Draw(screen *ebiten.Image) {
	g.game.Draw(screen)
}

func (g *closer) Layout(w, h int) (int, int) {
	return g.game.Layout(w, h)
}
