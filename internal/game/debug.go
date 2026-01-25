package game

// GetPlayers returns the internal players map.
// Intended for testing and debugging.
func (g *Game) GetPlayers() map[int]*Player {
	return g.players
}
