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

	town := NewWorldMap("town")
	field := NewWorldMap("field")
	dungeon := NewWorldMap("dungeon")

	g.maps["town"] = town
	g.maps["field"] = field
	g.maps["dungeon"] = dungeon

	town.Portals = append(town.Portals, &Portal{
		X:         750,
		Y:         300,
		Radius:    30,
		TargetMap: field,
		TargetX:   50,
		TargetY:   300,
	})

	field.Portals = append(field.Portals, &Portal{
		X:         50,
		Y:         300,
		Radius:    30,
		TargetMap: town,
		TargetX:   750,
		TargetY:   300,
	}, &Portal{
		X:         750,
		Y:         300,
		Radius:    30,
		TargetMap: dungeon,
		TargetX:   50,
		TargetY:   300,
	})

	dungeon.Portals = append(dungeon.Portals, &Portal{
		X:         50,
		Y:         300,
		Radius:    30,
		TargetMap: field,
		TargetX:   750,
		TargetY:   300,
	})

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
		g.checkPortalCollisions(p)
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
		m.UpdateMonsters(mapPlayers)
		m.UpdateItems(mapPlayers)
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
			snap.Projectiles = append(snap.Projectiles, &Entity{ID: proj.ID, X: proj.X, Y: proj.Y, Type: int(proj.Type)})
		}

		data, err := json.Marshal(snap)
		if err == nil {
			for _, p := range mapPlayers {
				p.Send(data)
			}
		}
	}
}

func (g *Game) checkPortalCollisions(p *Player) {
	if time.Since(p.LastPortalUse) < 2*time.Second {
		return
	}

	m, ok := g.maps[p.MapID]
	if !ok {
		return
	}

	for _, portal := range m.Portals {
		dx := p.X - portal.X
		dy := p.Y - portal.Y
		if dx*dx+dy*dy < portal.Radius*portal.Radius {
			g.switchMap(p, portal.TargetMap.ID, portal.TargetX, portal.TargetY)
			return
		}
	}
}

func (g *Game) switchMap(p *Player, targetMap string, targetX, targetY float64) {
	p.MapID = targetMap
	p.X = targetX
	p.Y = targetY
	p.LastPortalUse = time.Now()

	var portals []PortalData
	if m, ok := g.maps[targetMap]; ok {
		for _, p := range m.Portals {
			portals = append(portals, PortalData{
				X:      p.X,
				Y:      p.Y,
				Radius: p.Radius,
				Target: p.TargetMap.ID,
			})
		}
	}

	p.SendJSON(MsgMapSwitch{
		Type:    "MAP_SWITCH",
		Map:     targetMap,
		X:       targetX,
		Y:       targetY,
		Portals: portals,
	})

	if m, ok := g.maps[targetMap]; ok {
		for _, item := range m.Items {
			p.SendJSON(MsgItemSpawn{
				Type:     "ITEM_SPAWN",
				ID:       item.ID,
				ItemType: int(item.Type),
				X:        item.X,
				Y:        item.Y,
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

	// Fill inventory for testing
	for i := 0; i < 20; i++ {
		var item *Item
		if i < 10 {
			pType := ProjectileType(1 + rand.Intn(3))
			item = &Item{
				ID:             -1000 - i,
				Type:           ItemTypeWeapon,
				Name:           fmt.Sprintf("Test Sword %d", i),
				Attack:         10 + i,
				ProjectileType: pType,
			}
		} else {
			item = &Item{
				ID:      -1000 - i,
				Type:    ItemTypeArmor,
				Name:    fmt.Sprintf("Test Shield %d", i),
				Defense: 5 + (i - 10),
			}
		}
		p.Inventory = append(p.Inventory, item)
	}
	p.SendInventory()

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

	// Send initial map info (portals)
	if m, ok := g.maps[p.MapID]; ok {
		var portals []PortalData
		for _, por := range m.Portals {
			portals = append(portals, PortalData{
				X:      por.X,
				Y:      por.Y,
				Radius: por.Radius,
				Target: por.TargetMap.ID,
			})
		}

		// Reuse MsgMapSwitch to set initial map and portals
		p.SendJSON(MsgMapSwitch{
			Type:    "MAP_SWITCH",
			Map:     p.MapID,
			X:       p.X,
			Y:       p.Y,
			Portals: portals,
		})

		for _, item := range m.Items {
			p.SendJSON(MsgItemSpawn{
				Type:     "ITEM_SPAWN",
				ID:       item.ID,
				ItemType: int(item.Type),
				X:        item.X,
				Y:        item.Y,
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
