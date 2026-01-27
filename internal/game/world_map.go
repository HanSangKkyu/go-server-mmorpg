package game

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type Portal struct {
	X, Y      float64
	Radius    float64
	TargetMap *WorldMap
	TargetX   float64
	TargetY   float64
}

type WorldMap struct {
	ID          string
	Items       map[int]*Item
	Monsters    map[int]*Monster
	Projectiles map[int]*Projectile
	Portals     []*Portal

	Width  float64
	Height float64

	lastItemID int
	lastMonID  int
	lastProjID int

	lock sync.RWMutex
}

func NewWorldMap(id string) *WorldMap {
	return &WorldMap{
		ID:          id,
		Items:       make(map[int]*Item),
		Monsters:    make(map[int]*Monster),
		Projectiles: make(map[int]*Projectile),
		Portals:     make([]*Portal, 0),
		Width:       800,
		Height:      600,
	}
}

func (m *WorldMap) UpdateProjectiles() {
	m.lock.Lock()
	defer m.lock.Unlock()

	const speed = 10.0
	idsToRemove := []int{}

	for id, p := range m.Projectiles {
		p.X += p.VX * speed
		p.Y += p.VY * speed

		if p.X < -50 || p.X > m.Width+50 || p.Y < -50 || p.Y > m.Height+50 {
			idsToRemove = append(idsToRemove, id)
		}
	}

	for _, id := range idsToRemove {
		delete(m.Projectiles, id)
	}
}

func (m *WorldMap) UpdateItems(players []*Player) {
	m.lock.Lock()
	defer m.lock.Unlock()

	itemsToRemove := []int{}
	now := time.Now()

	for id, item := range m.Items {
		if now.Sub(item.CreatedAt) > 2*time.Minute {
			itemsToRemove = append(itemsToRemove, id)
		}
	}

	for _, id := range itemsToRemove {
		delete(m.Items, id)
		msg := MsgItemRemove{
			Type: "ITEM_REMOVE",
			ID:   id,
		}
		m.broadcastJSON(msg, players)
	}
}

func (m *WorldMap) UpdatePlayerShooting(players []*Player) {
	m.lock.Lock()
	defer m.lock.Unlock()

	now := time.Now()
	for _, p := range players {
		if now.Sub(p.LastShoot) > time.Millisecond*500 {
			var target *Monster
			minDist := math.MaxFloat64

			for _, mon := range m.Monsters {
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
			m.lastProjID++

			projType := 0
			for _, item := range p.Equipment {
				if item != nil && item.ProjectileType > 0 {
					projType = item.ProjectileType
				}
			}

			proj := &Projectile{
				ID:      m.lastProjID,
				OwnerID: p.ID,
				X:       p.X,
				Y:       p.Y,
				VX:      vx,
				VY:      vy,
				Type:    projType,
			}
			m.Projectiles[proj.ID] = proj
		}
	}
}

func (m *WorldMap) SpawnMonster() {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(m.Monsters) >= 10 {
		return
	}

	m.lastMonID++
	mon := &Monster{
		ID:    m.lastMonID,
		X:     50 + rand.Float64()*(m.Width-100),
		Y:     50 + rand.Float64()*(m.Height-100),
		Type:  MonsterType(rand.Intn(3)),
		HP:    50,
		MaxHP: 50,
	}
	m.Monsters[mon.ID] = mon
}

func (m *WorldMap) UpdateMonsters(players []*Player) {
	m.lock.Lock()
	defer m.lock.Unlock()

	const monsterSpeed = 2.0

	for _, mon := range m.Monsters {
		var target *Player
		minDistSq := math.MaxFloat64

		for _, p := range players {
			dx := p.X - mon.X
			dy := p.Y - mon.Y
			distSq := dx*dx + dy*dy
			if distSq < minDistSq {
				minDistSq = distSq
				target = p
			}
		}

		vx, vy := 0.0, 0.0

		if target != nil {
			dx := target.X - mon.X
			dy := target.Y - mon.Y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > monsterSpeed {
				vx = (dx / dist) * monsterSpeed
				vy = (dy / dist) * monsterSpeed
			}
		}

		for _, other := range m.Monsters {
			if mon == other {
				continue
			}
			dx := mon.X - other.X
			dy := mon.Y - other.Y
			distSq := dx*dx + dy*dy
			const collisionDistance = 40.0

			if distSq < collisionDistance*collisionDistance && distSq > 0 {
				dist := math.Sqrt(distSq)
				push := (collisionDistance - dist) / dist
				vx += dx * push * 0.1
				vy += dy * push * 0.1
			}
		}

		mon.X += vx
		mon.Y += vy
	}
}

func (m *WorldMap) CheckCollisions(players []*Player) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Build a map for faster player lookup
	playerMap := make(map[int]*Player)
	for _, p := range players {
		playerMap[p.ID] = p
	}

	projToRemove := make(map[int]bool)
	monstersToKill := []int{}

	for pid, proj := range m.Projectiles {
		for mid, mon := range m.Monsters {
			dx := proj.X - mon.X
			dy := proj.Y - mon.Y
			if dx*dx+dy*dy < 400 {
				projToRemove[pid] = true

				damage := 10
				if owner, ok := playerMap[proj.OwnerID]; ok {
					damage = owner.Attack
				}
				mon.HP -= damage

				if mon.HP <= 0 {
					monstersToKill = append(monstersToKill, mid)
					m.spawnItemAt(mon.X, mon.Y, players)
				}

				break
			}
		}
	}

	for pid := range projToRemove {
		delete(m.Projectiles, pid)
	}
	for _, mid := range monstersToKill {
		delete(m.Monsters, mid)
	}

	const collectRadius = 15.0
	for _, p := range players {
		for _, item := range m.Items {
			dx := p.X - item.X
			dy := p.Y - item.Y
			if dx*dx+dy*dy < collectRadius*collectRadius {
				m.collectItem(p, item, players)
			}
		}
	}
}

func (m *WorldMap) spawnItemAt(x, y float64, players []*Player) {
	m.lastItemID++

	randVal := rand.Float64()
	var iType ItemType
	var name string
	var atk, def, projType int

	if randVal < 0.5 {
		iType = ItemTypeGold
		name = "Gold"
	} else if randVal < 0.75 {
		iType = ItemTypeWeapon
		name = "Sword"
		atk = 5 + rand.Intn(10)
		projType = 1 + rand.Intn(2)
	} else {
		iType = ItemTypeArmor
		name = "Shield"
		def = 2 + rand.Intn(5)
	}

	item := &Item{
		ID:             m.lastItemID,
		X:              x,
		Y:              y,
		CreatedAt:      time.Now(),
		Type:           iType,
		Name:           name,
		Attack:         atk,
		Defense:        def,
		ProjectileType: projType,
	}
	m.Items[item.ID] = item

	msg := MsgItemSpawn{
		Type:     "ITEM_SPAWN",
		ID:       item.ID,
		ItemType: int(iType),
		X:        item.X,
		Y:        item.Y,
	}
	m.broadcastJSON(msg, players)
}

func (m *WorldMap) collectItem(p *Player, item *Item, players []*Player) {
	delete(m.Items, item.ID)

	if item.Type == ItemTypeGold {
		p.Gold += 100
		p.SendJSON(MsgGoldUpdate{
			Type:   "GOLD_UPDATE",
			Amount: p.Gold,
		})
	} else {
		p.Inventory = append(p.Inventory, item)
		p.SendInventory()
	}

	m.broadcastJSON(MsgItemRemove{
		Type: "ITEM_REMOVE",
		ID:   item.ID,
	}, players)
}

func (m *WorldMap) broadcastJSON(v interface{}, players []*Player) {
	for _, p := range players {
		p.SendJSON(v)
	}
}

func (m *WorldMap) AddProjectile(proj *Projectile) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Projectiles[proj.ID] = proj
}
