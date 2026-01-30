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
	ID            int
	MapID         string
	Conn          Connection
	X             float64
	Y             float64
	DirX          float64
	DirY          float64
	LastShoot     time.Time
	LastPortalUse time.Time
	Inventory     []*Item
	Equipment     [5]*Item

	HP      int
	MaxHP   int
	Attack  int
	Defense int
	Speed   float64
	Gold    int

	game *Game
}

func NewPlayer(id int, conn Connection, g *Game) *Player {
	return &Player{
		ID:        id,
		MapID:     "town",
		Conn:      conn,
		X:         400,
		Y:         300,
		DirX:      0,
		DirY:      1,
		Inventory: make([]*Item, 0),

		HP:      100,
		MaxHP:   100,
		Attack:  10,
		Defense: 0,
		Speed:   5.0,
		Gold:    0,
		game:    g,
	}
}

func (p *Player) Equip(itemID int, slot int) {
	if slot < 0 || slot >= 5 {
		return
	}

	var itemIdx int = -1
	for i, it := range p.Inventory {
		if it.ID == itemID {
			itemIdx = i
			break
		}
	}

	if itemIdx == -1 {
		return
	}

	item := p.Inventory[itemIdx]

	if p.Equipment[slot] != nil {
		p.Unequip(slot)
	}

	p.Inventory = append(p.Inventory[:itemIdx], p.Inventory[itemIdx+1:]...)

	p.Equipment[slot] = item
	p.RecalculateStats()
	p.SendInventory()
	p.SendEquipment()
}

func (p *Player) Unequip(slot int) {
	if slot < 0 || slot >= 5 {
		return
	}

	item := p.Equipment[slot]
	if item == nil {
		return
	}

	p.Equipment[slot] = nil
	p.Inventory = append(p.Inventory, item)
	p.RecalculateStats()
	p.SendInventory()
	p.SendEquipment()
}

func (p *Player) Sell(itemID int) {
	if p.game != nil {
		m, ok := p.game.maps[p.MapID]
		if ok {
			nearShop := false
			for _, npc := range m.NPCs {
				if npc.Type == NPCTypeShop {
					dx := p.X - npc.X
					dy := p.Y - npc.Y
					if dx*dx+dy*dy < 100*100 {
						nearShop = true
						break
					}
				}
			}
			if !nearShop {
				return
			}
		}
	}

	var itemIdx int = -1
	for i, it := range p.Inventory {
		if it.ID == itemID {
			itemIdx = i
			break
		}
	}

	if itemIdx == -1 {
		return
	}

	item := p.Inventory[itemIdx]

	// Price formula: 10 + (Atk + Def + Spd) * 5
	statsSum := item.Attack + item.Defense + int(item.Speed)
	price := 10 + statsSum*5

	p.Gold += price

	p.Inventory = append(p.Inventory[:itemIdx], p.Inventory[itemIdx+1:]...)

	p.SendInventory()
	p.SendJSON(MsgGoldUpdate{
		Type:   "GOLD_UPDATE",
		Amount: p.Gold,
	})
}

func (p *Player) RecalculateStats() {
	atk := 10
	def := 0
	spd := 5.0

	for _, item := range p.Equipment {
		if item != nil {
			atk += item.Attack
			def += item.Defense
			if item.Speed > 0 {
				spd += item.Speed
			}
		}
	}

	p.Attack = atk
	p.Defense = def
	p.Speed = spd

	p.SendJSON(MsgWelcome{
		Type:    "STATS",
		ID:      p.ID,
		HP:      p.HP,
		MaxHP:   p.MaxHP,
		Attack:  p.Attack,
		Defense: p.Defense,
		Speed:   p.Speed,
		Gold:    p.Gold,
	})
}

func (p *Player) SendInventory() {
	p.SendJSON(MsgInventory{
		Type:  "INVENTORY",
		Items: p.Inventory,
	})
}

func (p *Player) SendEquipment() {
	equipMap := make(map[int]*Item)
	for i, it := range p.Equipment {
		if it != nil {
			equipMap[i] = it
		}
	}
	p.SendJSON(MsgEquipment{
		Type:  "EQUIPMENT",
		Items: equipMap,
	})
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
