package game

import (
	"encoding/json"
	"fmt"
	"time"
)

// ListMarketItem lists an item from a player's inventory to the market
func (g *Game) ListMarketItem(p *Player, itemID int, price int) {
	if price <= 0 {
		return
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	var item *Item
	itemIdx := -1
	for i, it := range p.Inventory {
		if it.ID == itemID {
			item = it
			itemIdx = i
			break
		}
	}

	if itemIdx == -1 {
		return
	}

	p.Inventory = append(p.Inventory[:itemIdx], p.Inventory[itemIdx+1:]...)

	g.lastMarketID++
	marketItem := &MarketItem{
		ID:         g.lastMarketID,
		SellerID:   p.ID,
		SellerName: fmt.Sprintf("Player %d", p.ID),
		Item:       item,
		Price:      price,
		CreatedAt:  time.Now(),
	}
	g.market[marketItem.ID] = marketItem

	p.SendInventory()

	g.broadcastMarket()
}

// BuyMarketItem handles buying an item from the market
func (g *Game) BuyMarketItem(buyer *Player, marketID int) {
	g.lock.Lock()
	defer g.lock.Unlock()

	mItem, ok := g.market[marketID]
	if !ok {
		return
	}

	if mItem.SellerID == buyer.ID {
		delete(g.market, marketID)
		buyer.Inventory = append(buyer.Inventory, mItem.Item)
		buyer.SendInventory()
		g.broadcastMarket()
		return
	}

	if buyer.Gold < mItem.Price {
		return
	}

	buyer.Gold -= mItem.Price

	if seller, ok := g.players[mItem.SellerID]; ok {
		seller.Gold += mItem.Price
		seller.SendJSON(MsgGoldUpdate{
			Type:   "GOLD_UPDATE",
			Amount: seller.Gold,
		})
	}

	buyer.Inventory = append(buyer.Inventory, mItem.Item)
	delete(g.market, marketID)

	buyer.SendJSON(MsgGoldUpdate{
		Type:   "GOLD_UPDATE",
		Amount: buyer.Gold,
	})
	buyer.SendInventory()
	g.broadcastMarket()
}

// SendMarket sends the current market state to a specific player
func (g *Game) SendMarket(p *Player) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	var items []*MarketItem
	for _, it := range g.market {
		items = append(items, it)
	}

	p.SendJSON(MsgMarketUpdate{
		Type:  "MARKET_UPDATE",
		Items: items,
	})
}

func (g *Game) broadcastMarket() {
	var items []*MarketItem
	for _, it := range g.market {
		items = append(items, it)
	}

	msg := MsgMarketUpdate{
		Type:  "MARKET_UPDATE",
		Items: items,
	}

	data, err := json.Marshal(msg)
	if err == nil {
		for _, p := range g.players {
			p.Send(data)
		}
	}
}
