package game

import "time"

type ItemType int

const (
	ItemTypeGold ItemType = iota
	ItemTypeWeapon
	ItemTypeArmor
)

type Item struct {
	ID        int
	Type      ItemType
	Name      string
	X         float64
	Y         float64
	CreatedAt time.Time

	// Stats
	Attack  int
	Defense int
	Speed   float64

	// Special
	ProjectileType int
}

type MonsterType int

const (
	MonsterTypeWater MonsterType = iota
	MonsterTypeFire
	MonsterTypeGrass
)

type Monster struct {
	ID    int
	X     float64
	Y     float64
	Type  MonsterType
	HP    int
	MaxHP int
}

type Projectile struct {
	ID      int
	OwnerID int
	X       float64
	Y       float64
	VX      float64
	VY      float64
	Type    int // 0: Default, 1: Fire, 2: Ice, etc.
}
