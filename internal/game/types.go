package game

import "time"

type Item struct {
	ID        int
	X         float64
	Y         float64
	CreatedAt time.Time
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
}
