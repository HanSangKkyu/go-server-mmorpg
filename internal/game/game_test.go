package game

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	g := NewGame()
	if g == nil {
		t.Fatal("NewGame returned nil")
	}
	if len(g.players) != 0 {
		t.Errorf("Expected 0 players, got %d", len(g.players))
	}
}

func TestGame_AddRemovePlayer(t *testing.T) {
	g := NewGame()
	
	p := g.AddPlayer(nil)
	if p == nil {
		t.Fatal("AddPlayer returned nil")
	}
	if len(g.players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(g.players))
	}
	if p.ID != 1 {
		t.Errorf("Expected player ID 1, got %d", p.ID)
	}

	g.RemovePlayer(p.ID)
	if len(g.players) != 0 {
		t.Errorf("Expected 0 players after remove, got %d", len(g.players))
	}
}
