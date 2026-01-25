package game_test

import (
	"mmorpg/internal/game"
	"testing"
)

func TestNewPlayer(t *testing.T) {
	p := game.NewPlayer(1, nil)
	if p.ID != 1 {
		t.Errorf("Expected player ID 1, got %d", p.ID)
	}
	if p.X != 0 || p.Y != 0 {
		t.Errorf("Expected initial position (0,0), got (%.2f,%.2f)", p.X, p.Y)
	}
}

func TestPlayer_Move(t *testing.T) {
	p := game.NewPlayer(1, nil)
	p.Move(10.5, 20.0)

	if p.X != 10.5 {
		t.Errorf("Expected X 10.5, got %.2f", p.X)
	}
	if p.Y != 20.0 {
		t.Errorf("Expected Y 20.0, got %.2f", p.Y)
	}
}
