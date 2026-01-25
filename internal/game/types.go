package game

type Item struct {
	ID int
	X  float64
	Y  float64
}

type Monster struct {
	ID    int
	X     float64
	Y     float64
	Type  int // 0: Water, 1: Fire, 2: Grass
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
