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
	Type  int     `json:"type"`
	HP    int     `json:"hp,omitempty"`
	MaxHP int     `json:"max_hp,omitempty"`
}

// MsgSnap - Server -> Client
type MsgSnap struct {
	Type        string    `json:"type"`
	Players     []*Entity `json:"players"`
	Monsters    []*Entity `json:"monsters"`
	Projectiles []*Entity `json:"projectiles"`
}

// MsgItemSpawn - Server -> Client
type MsgItemSpawn struct {
	Type     string  `json:"type"`
	ID       int     `json:"id"`
	ItemType int     `json:"item_type"` // 0: Gold, 1: Weapon...
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
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

type PortalData struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Radius float64 `json:"radius"`
	Target string  `json:"target"`
}

// MsgMapSwitch - Server -> Client
type MsgMapSwitch struct {
	Type    string       `json:"type"`
	Map     string       `json:"map"`
	X       float64      `json:"x"`
	Y       float64      `json:"y"`
	Portals []PortalData `json:"portals"`
}

// MsgInventory - Server -> Client
type MsgInventory struct {
	Type  string  `json:"type"`
	Items []*Item `json:"items"`
}

// MsgEquipment - Server -> Client
type MsgEquipment struct {
	Type  string        `json:"type"`
	Items map[int]*Item `json:"items"` // Slot -> Item
}

// MsgEquip - Client -> Server
type MsgEquip struct {
	Type   string `json:"type"`
	ItemID int    `json:"item_id"`
	Slot   int    `json:"slot"`
}

// MsgUnequip - Client -> Server
type MsgUnequip struct {
	Type string `json:"type"`
	Slot int    `json:"slot"`
}

// MsgMove - Client -> Server
type MsgMove struct {
	Type string  `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}
