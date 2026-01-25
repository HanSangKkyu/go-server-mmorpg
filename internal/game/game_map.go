package game

import (
	"encoding/json"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Map struct {
	Name        string
	players     map[int]*Player
	items       map[int]*Item
	monsters    map[int]*Monster
	projectiles map[int]*Projectile
	portals     []*Portal

	lock sync.RWMutex
	game *Game // Reference back to global game context
}

func NewMap(name string, g *Game) *Map {
	return &Map{
		Name:        name,
		game:        g,
		players:     make(map[int]*Player),
		items:       make(map[int]*Item),
		monsters:    make(map[int]*Monster),
		projectiles: make(map[int]*Projectile),
		portals:     make([]*Portal, 0),
	}
}

func (m *Map) Update() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.updateProjectiles()
	m.updatePlayerShooting()
	m.checkCollisions()

	if len(m.players) == 0 {
		return
	}

	// Create JSON Snapshot
	snap := MsgSnap{
		Type:        "SNAP",
		Players:     make([]*Entity, 0, len(m.players)),
		Monsters:    make([]*Entity, 0, len(m.monsters)),
		Projectiles: make([]*Entity, 0, len(m.projectiles)),
		Portals:     make([]*PortalEntity, 0, len(m.portals)),
	}

	for _, p := range m.players {
		snap.Players = append(snap.Players, &Entity{ID: p.ID, X: p.X, Y: p.Y})
	}
	for _, mon := range m.monsters {
		snap.Monsters = append(snap.Monsters, &Entity{
			ID:    mon.ID,
			X:     mon.X,
			Y:     mon.Y,
			Type:  mon.Type,
			HP:    mon.HP,
			MaxHP: mon.MaxHP,
		})
	}
	for _, proj := range m.projectiles {
		snap.Projectiles = append(snap.Projectiles, &Entity{ID: proj.ID, X: proj.X, Y: proj.Y})
	}
	for _, port := range m.portals {
		snap.Portals = append(snap.Portals, &PortalEntity{
			ID:        port.ID,
			X:         port.X,
			Y:         port.Y,
			TargetMap: port.TargetMap,
		})
	}

	// Broadcast
	data, err := json.Marshal(snap)
	if err == nil {
		for _, p := range m.players {
			p.Send(data)
		}
	}
}

func (m *Map) updateProjectiles() {
	const speed = 10.0
	idsToRemove := []int{}

	for id, p := range m.projectiles {
		p.X += p.VX * speed
		p.Y += p.VY * speed
		if p.X < -50 || p.X > 850 || p.Y < -50 || p.Y > 650 {
			idsToRemove = append(idsToRemove, id)
		}
	}
	for _, id := range idsToRemove {
		delete(m.projectiles, id)
	}
}

func (m *Map) updatePlayerShooting() {
	now := time.Now()
	for _, p := range m.players {
		if now.Sub(p.LastShoot) > time.Millisecond*500 {
			var target *Monster
			minDist := math.MaxFloat64

			for _, mon := range m.monsters {
				dx := mon.X - p.X
				dy := mon.Y - p.Y
				dist := dx*dx + dy*dy
				if dist < minDist {
					minDist = dist
					target = mon
				}
			}

			vx, vy := p.DirX, p.DirY
			if target != nil {
				dx := target.X - p.X
				dy := target.Y - p.Y
				len := math.Sqrt(dx*dx + dy*dy)
				if len > 0 {
					vx = dx / len
					vy = dy / len
				}
			} else {
				continue
			}

			p.LastShoot = now

			// Generate Global ID from Game
			id := m.game.GenerateProjID()
			proj := &Projectile{
				ID:      id,
				OwnerID: p.ID,
				X:       p.X,
				Y:       p.Y,
				VX:      vx,
				VY:      vy,
			}
			m.projectiles[proj.ID] = proj
		}
	}
}

func (m *Map) checkCollisions() {
	// 1. Projectiles vs Monsters
	projToRemove := make(map[int]bool)
	monstersToKill := []int{}

	for pid, proj := range m.projectiles {
		for mid, mon := range m.monsters {
			dx := proj.X - mon.X
			dy := proj.Y - mon.Y
			if dx*dx+dy*dy < 400 {
				projToRemove[pid] = true
				mon.HP -= 10
				if mon.HP <= 0 {
					monstersToKill = append(monstersToKill, mid)
					m.spawnItemAt(mon.X, mon.Y)
				}
				break
			}
		}
	}
	for pid := range projToRemove {
		delete(m.projectiles, pid)
	}
	for _, mid := range monstersToKill {
		delete(m.monsters, mid)
	}

	// 2. Player vs Items
	const collectRadius = 15.0
	for _, p := range m.players {
		for _, item := range m.items {
			dx := p.X - item.X
			dy := p.Y - item.Y
			if dx*dx+dy*dy < collectRadius*collectRadius {
				m.collectItem(p, item)
			}
		}
	}

	// 3. Player vs Portals
	const portalRadius = 30.0
	for _, p := range m.players {
		for _, port := range m.portals {
			dx := p.X - port.X
			dy := p.Y - port.Y
			if dx*dx+dy*dy < portalRadius*portalRadius {
				// Request Map Switch
				// We call Game.SwitchMap via goroutine to avoid deadlock (Game lock vs Map lock)
				// Or better: Use channel or make SwitchMap safe.
				// For simplicity: async call.
				go m.game.SwitchMap(p.ID, port.TargetMap, port.TargetX, port.TargetY)
			}
		}
	}
}

func (m *Map) spawnItemAt(x, y float64) {
	id := m.game.GenerateItemID()
	item := &Item{ID: id, X: x, Y: y}
	m.items[id] = item

	msg := MsgItemSpawn{Type: "ITEM_SPAWN", ID: id, X: x, Y: y}
	m.broadcastJSON(msg)
}

func (m *Map) collectItem(p *Player, item *Item) {
	delete(m.items, item.ID)
	p.Gold += 100
	p.SendJSON(MsgGoldUpdate{Type: "GOLD_UPDATE", Amount: p.Gold})
	m.broadcastJSON(MsgItemRemove{Type: "ITEM_REMOVE", ID: item.ID})
}

func (m *Map) SpawnMonster() {
	m.lock.Lock()
	defer m.lock.Unlock()

	id := m.game.GenerateMonID()
	mon := &Monster{
		ID:    id,
		X:     50 + rand.Float64()*700,
		Y:     50 + rand.Float64()*500,
		Type:  rand.Intn(3),
		HP:    50,
		MaxHP: 50,
	}
	m.monsters[id] = mon
}

func (m *Map) AddPlayer(p *Player) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.players[p.ID] = p

	// Send existing items to new player
	for _, item := range m.items {
		p.SendJSON(MsgItemSpawn{Type: "ITEM_SPAWN", ID: item.ID, X: item.X, Y: item.Y})
	}

	// Send existing portals
	// Actually portals are in SNAP, but initial render might need them?
	// SNAP comes 33ms later, so fine.
}

func (m *Map) RemovePlayer(id int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.players, id)
	m.broadcastJSON(MsgLeave{Type: "LEAVE", ID: id})
}

func (m *Map) broadcastJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err == nil {
		for _, p := range m.players {
			p.Send(data)
		}
	}
}
