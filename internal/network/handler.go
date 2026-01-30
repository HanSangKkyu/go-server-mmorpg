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
	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(text), &msg); err == nil {
		msgType, ok := msg["type"].(string)
		if !ok {
			return
		}

		switch msgType {
		case "MOVE":
			var move game.MsgMove
			if err := json.Unmarshal([]byte(text), &move); err == nil {
				player.Move(move.X, move.Y)
			}
		case "EQUIP":
			var equip game.MsgEquip
			if err := json.Unmarshal([]byte(text), &equip); err == nil {
				player.Equip(equip.ItemID, equip.Slot)
			}
		case "UNEQUIP":
			var unequip game.MsgUnequip
			if err := json.Unmarshal([]byte(text), &unequip); err == nil {
				player.Unequip(unequip.Slot)
			}
		case "SELL":
			var sell game.MsgSell
			if err := json.Unmarshal([]byte(text), &sell); err == nil {
				player.Sell(sell.ItemID)
			}
		case "MARKET_LIST":
			var list game.MsgMarketList
			if err := json.Unmarshal([]byte(text), &list); err == nil {
				if player.Game() != nil {
					player.Game().ListMarketItem(player, list.ItemID, list.Price)
				}
			}
		case "MARKET_BUY":
			var buy game.MsgMarketBuy
			if err := json.Unmarshal([]byte(text), &buy); err == nil {
				if player.Game() != nil {
					player.Game().BuyMarketItem(player, buy.MarketID)
				}
			}
		}
		return
	}

	// Fallback for debugging/legacy (optional)
	// fmt.Printf("Unknown command from %d: %s\n", player.ID, text)
}
