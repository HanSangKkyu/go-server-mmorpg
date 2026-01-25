package network

import (
	"bufio"
	"mmorpg/internal/game"
	"net"
	"testing"
	"time"
)

func TestHandleConnection(t *testing.T) {
	// Use net.Pipe to simulate a connection without opening a real port
	serverConn, clientConn := net.Pipe()

	g := game.NewGame()
	s := NewServer(":0", g)

	s.wg.Add(1)
	
	go s.handleConnection(serverConn)

	// Reader for client to see responses
	reader := bufio.NewReader(clientConn)

	// Test 1: Send MOVE command
	_, err := clientConn.Write([]byte("MOVE 10.5 20.0\n"))
	if err != nil {
		t.Fatalf("Failed to write to pipe: %v", err)
	}

	// Read response
	resp, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	expected := "Moved to 10.50, 20.00\n"
	if resp != expected {
		t.Errorf("Expected %q, got %q", expected, resp)
	}

	// Test 2: Verify game state updated
	time.Sleep(10 * time.Millisecond)
	
	_, err = clientConn.Write([]byte("JUMP 5\n"))
	if err != nil {
		t.Fatalf("Failed to write to pipe: %v", err)
	}
	resp, err = reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	expectedError := "Unknown command. Try MOVE x y or SAY msg\n"
	if resp != expectedError {
		t.Errorf("Expected %q, got %q", expectedError, resp)
	}

	// Clean up
	clientConn.Close()
	// serverConn will be closed by handleConnection when it detects EOF or error
}
