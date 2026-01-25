package game

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Game struct {
	players map[int]*Player
	maps    map[string]*WorldMap

	lock   sync.RWMutex
	lastID int
	quitch chan struct{}
}

func NewGame() *Game {
	g := &Game{
		players: make(map[int]*Player),
		maps:    make(map[string]*WorldMap),
		quitch:  make(chan struct{}),
	}

	g.maps["town"] = NewWorldMap("town")
	g.maps["field"] = NewWorldMap("field")
	g.maps["dungeon"] = NewWorldMap("dungeon")

	return g
}

func (g *Game) Start() {
	rand.Seed(time.Now().UnixNano())
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
			g.SpawnMonsters()
		}
	}
}

func (g *Game) SpawnMonsters() {
	for _, m := range g.maps {
		if m.ID == "town" {
			continue
		}
		m.SpawnMonster()
	}
}

func (g *Game) Update() {
	g.lock.Lock()
	defer g.lock.Unlock()

	playersByMap := make(map[string][]*Player)
	for _, p := range g.players {
		g.checkMapTransitions(p)
		playersByMap[p.MapID] = append(playersByMap[p.MapID], p)
	}

	if len(g.players) == 0 {
		return
	}

	for mapID, m := range g.maps {
		mapPlayers := playersByMap[mapID]
		if len(mapPlayers) == 0 && len(m.Projectiles) == 0 && len(m.Monsters) == 0 {
			continue
		}

		m.UpdateProjectiles()
		m.UpdatePlayerShooting(mapPlayers)
		m.CheckCollisions(mapPlayers)

		snap := MsgSnap{
			Type:        "SNAP",
			Players:     make([]*Entity, 0, len(mapPlayers)),
			Monsters:    make([]*Entity, 0, len(m.Monsters)),
			Projectiles: make([]*Entity, 0, len(m.Projectiles)),
		}

		for _, p := range mapPlayers {
			snap.Players = append(snap.Players, &Entity{ID: p.ID, X: p.X, Y: p.Y})
		}
		for _, mon := range m.Monsters {
			snap.Monsters = append(snap.Monsters, &Entity{
				ID:    mon.ID,
				X:     mon.X,
				Y:     mon.Y,
				Type:  int(mon.Type),
				HP:    mon.HP,
				MaxHP: mon.MaxHP,
			})
		}
		for _, proj := range m.Projectiles {
			snap.Projectiles = append(snap.Projectiles, &Entity{ID: proj.ID, X: proj.X, Y: proj.Y})
		}

		data, err := json.Marshal(snap)
		if err == nil {
			for _, p := range mapPlayers {
				p.Send(data)
			}
		}
	}
}

func (g *Game) checkMapTransitions(p *Player) {
	const edge = 20.0
	width := 800.0
	height := 600.0

	centerX := width / 2
	centerY := height / 2

	if p.MapID == "town" {
		if p.X > width-edge {
			g.switchMap(p, "field", centerX, centerY)
		}
	} else if p.MapID == "field" {
		if p.X < edge {
			g.switchMap(p, "town", centerX, centerY)
		} else if p.X > width-edge {
			g.switchMap(p, "dungeon", centerX, centerY)
		}
	} else if p.MapID == "dungeon" {
		if p.X < edge {
			g.switchMap(p, "field", centerX, centerY)
		}
	}
}

func (g *Game) switchMap(p *Player, targetMap string, targetX, targetY float64) {
	p.MapID = targetMap
	p.X = targetX
	p.Y = targetY

	p.SendJSON(MsgMapSwitch{
		Type: "MAP_SWITCH",
		Map:  targetMap,
		X:    targetX,
		Y:    targetY,
	})

	if m, ok := g.maps[targetMap]; ok {
		for _, item := range m.Items {
			p.SendJSON(MsgItemSpawn{
				Type: "ITEM_SPAWN",
				ID:   item.ID,
				X:    item.X,
				Y:    item.Y,
			})
		}
	}
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

	if m, ok := g.maps[p.MapID]; ok {
		for _, item := range m.Items {
			p.SendJSON(MsgItemSpawn{
				Type: "ITEM_SPAWN",
				ID:   item.ID,
				X:    item.X,
				Y:    item.Y,
			})
		}
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
