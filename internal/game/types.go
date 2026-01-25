package game

type Item struct {
	ID int
	X  float64
	Y  float64
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
