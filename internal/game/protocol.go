package game

// MsgWelcome - Server -> Client
type MsgWelcome struct {
	Type    string  `json:"type"`
	ID      int     `json:"id"`
	HP      int     `json:"hp"`
	MaxHP   int     `json:"max_hp"`
	Attack  int     `json:"attack"`
	Defense int     `json:"defense"`
	Speed   float64 `json:"speed"`
	Gold    int     `json:"gold"`
}

type Entity struct {
	ID    int     `json:"id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Type  int     `json:"type,omitempty"`
	HP    int     `json:"hp,omitempty"`
	MaxHP int     `json:"max_hp,omitempty"`
}

type PortalEntity struct {
	ID        int     `json:"id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	TargetMap string  `json:"target_map"`
}

// MsgSnap - Server -> Client
type MsgSnap struct {
	Type        string          `json:"type"`
	Players     []*Entity       `json:"players"`
	Monsters    []*Entity       `json:"monsters"`
	Projectiles []*Entity       `json:"projectiles"`
	Portals     []*PortalEntity `json:"portals"`
}

// MsgItemSpawn - Server -> Client
type MsgItemSpawn struct {
	Type string  `json:"type"`
	ID   int     `json:"id"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

// MsgItemRemove - Server -> Client
type MsgItemRemove struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

// MsgGoldUpdate - Server -> Client
type MsgGoldUpdate struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

// MsgLeave - Server -> Client
type MsgLeave struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

// MsgMapChange - Server -> Client
type MsgMapChange struct {
	Type    string `json:"type"`
	MapName string `json:"map_name"`
}

// MsgMove - Client -> Server
type MsgMove struct {
	Type string  `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}
