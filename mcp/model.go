package mcp

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	Timeout = 500 * time.Millisecond
	Alive   = 5 * time.Second
)

type State int

const (
	Start State = iota
	Record
	Exit
)

type SetupOptions struct {
	Draw bool
}

type Decorator interface {
	Setup(op SetupOptions)
	Update() error
	Draw(screen *ebiten.Image)
}
