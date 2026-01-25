package game

import (
	"fmt"
	"sync"
	"time"
)

type Game struct {
	maps       map[string]*Map
	players    map[int]*Player // Global player registry (ID -> Player)
	playerMaps map[int]string  // Track which map each player is in (ID -> MapName)

	lock       sync.RWMutex
	lastID     int
	lastItemID int
	lastMonID  int
	lastProjID int
	quitch     chan struct{}
}

func NewGame() *Game {
	g := &Game{
		maps:       make(map[string]*Map),
		players:    make(map[int]*Player),
		playerMaps: make(map[int]string),
		quitch:     make(chan struct{}),
	}

	// Initialize Maps
	// Map 1: "main"
	m1 := NewMap("main", g)
	m1.portals = append(m1.portals, &Portal{ID: 1, X: 750, Y: 300, TargetMap: "dungeon", TargetX: 100, TargetY: 300})
	g.maps["main"] = m1

	// Map 2: "dungeon"
	m2 := NewMap("dungeon", g)
	m2.portals = append(m2.portals, &Portal{ID: 2, X: 50, Y: 300, TargetMap: "main", TargetX: 700, TargetY: 300})
	g.maps["dungeon"] = m2

	return g
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
	// Update all maps
	for _, m := range g.maps {
		m.Update()
	}
}

func (g *Game) SpawnMonster() {
	// Spawn monsters in all maps
	for _, m := range g.maps {
		m.SpawnMonster()
	}
}

func (g *Game) AddPlayer(conn Connection) *Player {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.lastID++
	p := NewPlayer(g.lastID, conn)

	// Add to global registry
	g.players[p.ID] = p
	g.playerMaps[p.ID] = "main" // Start in main map

	fmt.Printf("Player joined: %d (Map: main)\n", p.ID)

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

	// Add to Map "main"
	if m, ok := g.maps["main"]; ok {
		m.AddPlayer(p)
	}

	return p
}

func (g *Game) RemovePlayer(id int) {
	g.lock.Lock()
	defer g.lock.Unlock()

	mapName, ok := g.playerMaps[id]
	if ok {
		if m, exists := g.maps[mapName]; exists {
			m.RemovePlayer(id)
		}
		delete(g.playerMaps, id)
	}
	delete(g.players, id)
	fmt.Printf("Player left: %d\n", id)
}

func (g *Game) SwitchMap(playerID int, targetMap string, x, y float64) {
	g.lock.Lock()
	defer g.lock.Unlock()

	p, ok := g.players[playerID]
	if !ok {
		return
	}

	oldMapName := g.playerMaps[playerID]
	if oldMapName == targetMap {
		return
	}

	// Remove from old map
	if oldMap, exists := g.maps[oldMapName]; exists {
		oldMap.RemovePlayer(playerID)
	}

	// Update Player State
	p.X = x
	p.Y = y
	g.playerMaps[playerID] = targetMap

	// Notify Client to clear screen (MAP_CHANGE)
	p.SendJSON(MsgMapChange{Type: "MAP_CHANGE", MapName: targetMap})

	// Add to new map
	if newMap, exists := g.maps[targetMap]; exists {
		newMap.AddPlayer(p)
	}

	fmt.Printf("Player %d switched to %s\n", playerID, targetMap)
}

// Global ID Generators
func (g *Game) GenerateItemID() int {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.lastItemID++
	return g.lastItemID
}

func (g *Game) GenerateMonID() int {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.lastMonID++
	return g.lastMonID
}

func (g *Game) GenerateProjID() int {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.lastProjID++
	return g.lastProjID
}

// GetPlayers for Debug/Test
func (g *Game) GetPlayers() map[int]*Player {
	g.lock.RLock()
	defer g.lock.RUnlock()
	// Return a copy or just the map (testing only)
	return g.players
}
