package game_test

import (
	"mmorpg/internal/game"
	"testing"
)

func TestNewGame(t *testing.T) {
	g := game.NewGame()
	if g == nil {
		t.Fatal("NewGame returned nil")
	}
	if len(g.GetPlayers()) != 0 {
		t.Errorf("Expected 0 players, got %d", len(g.GetPlayers()))
	}
}

func TestGame_AddRemovePlayer(t *testing.T) {
	g := game.NewGame()

	p := g.AddPlayer(nil)
	if p == nil {
		t.Fatal("AddPlayer returned nil")
	}
	if len(g.GetPlayers()) != 1 {
		t.Errorf("Expected 1 player, got %d", len(g.GetPlayers()))
	}
	if p.ID != 1 {
		t.Errorf("Expected player ID 1, got %d", p.ID)
	}

	g.RemovePlayer(p.ID)
	if len(g.GetPlayers()) != 0 {
		t.Errorf("Expected 0 players after remove, got %d", len(g.GetPlayers()))
	}
}
