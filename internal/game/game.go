package game

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Game struct {
	players     map[int]*Player
	items       map[int]*Item
	monsters    map[int]*Monster
	projectiles map[int]*Projectile

	lock       sync.RWMutex
	lastID     int
	lastItemID int
	lastMonID  int
	lastProjID int
	quitch     chan struct{}
}

func NewGame() *Game {
	return &Game{
		players:     make(map[int]*Player),
		items:       make(map[int]*Item),
		monsters:    make(map[int]*Monster),
		projectiles: make(map[int]*Projectile),
		quitch:      make(chan struct{}),
	}
}

func (g *Game) Start() {
	ticker := time.NewTicker(time.Millisecond * 33)
	defer ticker.Stop()

	monsterTicker := time.NewTicker(time.Second * 1)
	defer monsterTicker.Stop()

	for {
		select {
		case <-g.quitch:
			return
		case <-ticker.C:
			g.Update()
		case <-monsterTicker.C:
			g.SpawnMonster()
		}
	}
}

func (g *Game) Update() {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.updateProjectiles()
	g.updatePlayerShooting()
	g.checkCollisions()

	if len(g.players) == 0 {
		return
	}

	// Create JSON Snapshot
	snap := MsgSnap{
		Type:        "SNAP",
		Players:     make([]*Entity, 0, len(g.players)),
		Monsters:    make([]*Entity, 0, len(g.monsters)),
		Projectiles: make([]*Entity, 0, len(g.projectiles)),
	}

	for _, p := range g.players {
		snap.Players = append(snap.Players, &Entity{ID: p.ID, X: p.X, Y: p.Y})
	}
	for _, m := range g.monsters {
		snap.Monsters = append(snap.Monsters, &Entity{
			ID:    m.ID,
			X:     m.X,
			Y:     m.Y,
			Type:  m.Type,
			HP:    m.HP,
			MaxHP: m.MaxHP,
		})
	}
	for _, proj := range g.projectiles {
		snap.Projectiles = append(snap.Projectiles, &Entity{ID: proj.ID, X: proj.X, Y: proj.Y})
	}

	// Broadcast JSON
	data, err := json.Marshal(snap)
	if err == nil {
		for _, p := range g.players {
			p.Send(data)
		}
	}
}

func (g *Game) updateProjectiles() {
	const speed = 10.0
	const boundary = 1000.0

	idsToRemove := []int{}

	for id, p := range g.projectiles {
		p.X += p.VX * speed
		p.Y += p.VY * speed

		if p.X < -50 || p.X > 850 || p.Y < -50 || p.Y > 650 {
			idsToRemove = append(idsToRemove, id)
		}
	}

	for _, id := range idsToRemove {
		delete(g.projectiles, id)
	}
}

func (g *Game) updatePlayerShooting() {
	now := time.Now()
	for _, p := range g.players {
		if now.Sub(p.LastShoot) > time.Millisecond*500 {
			var target *Monster
			minDist := math.MaxFloat64

			for _, m := range g.monsters {
				dx := m.X - p.X
				dy := m.Y - p.Y
				dist := dx*dx + dy*dy
				if dist < minDist {
					minDist = dist
					target = m
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
			g.lastProjID++

			proj := &Projectile{
				ID:      g.lastProjID,
				OwnerID: p.ID,
				X:       p.X,
				Y:       p.Y,
				VX:      vx,
				VY:      vy,
			}
			g.projectiles[proj.ID] = proj
		}
	}
}

func (g *Game) checkCollisions() {
	projToRemove := make(map[int]bool)
	monstersToKill := []int{}

	for pid, proj := range g.projectiles {
		for mid, mon := range g.monsters {
			dx := proj.X - mon.X
			dy := proj.Y - mon.Y
			if dx*dx+dy*dy < 400 {
				projToRemove[pid] = true

				damage := 10
				mon.HP -= damage

				if mon.HP <= 0 {
					monstersToKill = append(monstersToKill, mid)
					g.spawnItemAt(mon.X, mon.Y)
				}

				break
			}
		}
	}

	for pid := range projToRemove {
		delete(g.projectiles, pid)
	}
	for _, mid := range monstersToKill {
		delete(g.monsters, mid)
	}

	const collectRadius = 15.0
	for _, p := range g.players {
		for _, item := range g.items {
			dx := p.X - item.X
			dy := p.Y - item.Y
			if dx*dx+dy*dy < collectRadius*collectRadius {
				g.collectItem(p, item)
			}
		}
	}
}

func (g *Game) SpawnMonster() {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.lastMonID++
	m := &Monster{
		ID:    g.lastMonID,
		X:     50 + rand.Float64()*700,
		Y:     50 + rand.Float64()*500,
		Type:  rand.Intn(3),
		HP:    50,
		MaxHP: 50,
	}
	g.monsters[m.ID] = m
}

func (g *Game) spawnItemAt(x, y float64) {
	g.lastItemID++
	item := &Item{
		ID: g.lastItemID,
		X:  x,
		Y:  y,
	}
	g.items[item.ID] = item

	msg := MsgItemSpawn{
		Type: "ITEM_SPAWN",
		ID:   item.ID,
		X:    item.X,
		Y:    item.Y,
	}

	// Use SendJSON via marshaling helper locally or just marshal here
	// Since we are inside Game, we iterate players.
	// We can use a helper broadcastJSON
	g.broadcastJSON(msg)
}

func (g *Game) collectItem(p *Player, item *Item) {
	delete(g.items, item.ID)
	p.Gold += 100

	p.SendJSON(MsgGoldUpdate{
		Type:   "GOLD_UPDATE",
		Amount: p.Gold,
	})

	g.broadcastJSON(MsgItemRemove{
		Type: "ITEM_REMOVE",
		ID:   item.ID,
	})
}

func (g *Game) AddPlayer(conn Connection) *Player {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.lastID++
	p := NewPlayer(g.lastID, conn)
	g.players[p.ID] = p
	fmt.Printf("Player joined: %d\n", p.ID)

	p.SendJSON(MsgWelcome{
		Type:    "WELCOME",
		ID:      p.ID,
		HP:      p.HP,
		MaxHP:   p.MaxHP,
		Attack:  p.Attack,
		Defense: p.Defense,
		Speed:   p.Speed,
		Gold:    p.Gold,
	})

	for _, item := range g.items {
		p.SendJSON(MsgItemSpawn{
			Type: "ITEM_SPAWN",
			ID:   item.ID,
			X:    item.X,
			Y:    item.Y,
		})
	}

	return p
}

func (g *Game) RemovePlayer(id int) {
	g.lock.Lock()
	defer g.lock.Unlock()

	delete(g.players, id)
	fmt.Printf("Player left: %d\n", id)

	g.broadcastJSON(MsgLeave{
		Type: "LEAVE",
		ID:   id,
	})
}

func (g *Game) broadcastJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err == nil {
		for _, p := range g.players {
			p.Send(data)
		}
	}
}
