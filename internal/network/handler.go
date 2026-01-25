package network

import (
	"encoding/json"
	"mmorpg/internal/game"
	"strings"
)

// HandleCommand processes a single line of text or JSON from a player
func HandleCommand(player *game.Player, text string) {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return
	}

	// Try parsing as JSON first
	var move game.MsgMove
	if err := json.Unmarshal([]byte(text), &move); err == nil {
		if move.Type == "MOVE" {
			player.Move(move.X, move.Y)
			return
		}
	}

	// Fallback for debugging/legacy (optional)
	// fmt.Printf("Unknown command from %d: %s\n", player.ID, text)
}
