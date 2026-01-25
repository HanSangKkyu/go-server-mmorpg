package game

import (
	"encoding/json"
	"fmt"
	"time"
)

type Connection interface {
	Write([]byte) (int, error)
	Close() error
}

type Player struct {
	ID        int
	Conn      Connection
	X         float64
	Y         float64
	DirX      float64
	DirY      float64
	LastShoot time.Time
	Inventory []*Item

	HP      int
	MaxHP   int
	Attack  int
	Defense int
	Speed   float64
	Gold    int
}

func NewPlayer(id int, conn Connection) *Player {
	return &Player{
		ID:        id,
		Conn:      conn,
		X:         0,
		Y:         0,
		DirX:      0,
		DirY:      1,
		Inventory: make([]*Item, 0),

		HP:      100,
		MaxHP:   100,
		Attack:  10,
		Defense: 0,
		Speed:   5.0,
		Gold:    0,
	}
}

func (p *Player) Move(x, y float64) {
	dx := x - p.X
	dy := y - p.Y
	if dx != 0 || dy != 0 {
		p.DirX = dx
		p.DirY = dy
	}
	p.X = x
	p.Y = y
}

func (p *Player) Send(msg []byte) {
	if p.Conn != nil {
		p.Conn.Write(msg)
	}
}

func (p *Player) SendJSON(v interface{}) {
	if p.Conn != nil {
		data, err := json.Marshal(v)
		if err == nil {
			p.Conn.Write(data)
		}
	}
}

func (p *Player) String() string {
	return fmt.Sprintf("Player %d [%.2f, %.2f] Inv: %d", p.ID, p.X, p.Y, len(p.Inventory))
}
