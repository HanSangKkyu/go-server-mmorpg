package network_test

import (
	"bufio"
	"encoding/json"
	"mmorpg/internal/game"
	"mmorpg/internal/network"
	"net"
	"testing"
	"time"
)

func TestHandleConnection(t *testing.T) {
	// Use net.Pipe to simulate a connection without opening a real port
	serverConn, clientConn := net.Pipe()

	g := game.NewGame()
	s := network.NewServer(":0", g)

	s.GetWG().Add(1)

	go s.ExportedHandleConnection(serverConn)

	// Consume WELCOME and ITEM_SPAWN messages to prevent deadlock
	// Since net.Pipe is unbuffered, we must read what server sends immediately
	go func() {
		reader := bufio.NewReader(clientConn)
		for {
			_, err := reader.ReadString('\n')
			if err != nil {
				return
			}
		}
	}()

	// Test 1: Send MOVE command (JSON)
	moveCmd := game.MsgMove{Type: "MOVE", X: 10.5, Y: 20.0}
	data, _ := json.Marshal(moveCmd)
	_, err := clientConn.Write(append(data, '\n'))
	if err != nil {
		t.Fatalf("Failed to write to pipe: %v", err)
	}

	// We can't easily verify the response because it's a SNAP broadcast which happens on TICK.
	// The immediate response logic was removed/refactored.
	// Current server doesn't echo "Moved to...".
	// It relies on SNAP broadcast from game loop.
	// So we can only verify if game state updated.

	time.Sleep(100 * time.Millisecond) // Wait for processing

	// Verify directly on player state if we had access to the player instance.
	// Since we don't easily get the player instance from here (it's inside server),
	// we can check the Game state directly via Exported getter.

	players := g.GetPlayers()
	if len(players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(players))
	}
	// Player ID starts at 1
	p := players[1]
	if p == nil {
		t.Fatal("Player 1 not found")
	}

	if p.X != 10.5 || p.Y != 20.0 {
		t.Errorf("Expected pos (10.5, 20.0), got (%.2f, %.2f)", p.X, p.Y)
	}

	// Clean up
	clientConn.Close()
}
