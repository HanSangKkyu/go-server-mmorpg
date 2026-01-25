package network

import (
	"log"
	"mmorpg/internal/game"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSConnection wraps websocket.Conn to satisfy game.Connection interface
type WSConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *WSConnection) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	err := w.conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *WSConnection) Close() error {
	return w.conn.Close()
}

type WSServer struct {
	game *game.Game
}

func NewWSServer(g *game.Game) *WSServer {
	return &WSServer{game: g}
}

func (s *WSServer) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	wsConn := &WSConnection{conn: conn}
	player := s.game.AddPlayer(wsConn)
	defer s.game.RemovePlayer(player.ID)
	defer wsConn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Player %d disconnected: %v", player.ID, err)
			break
		}

		HandleCommand(player, string(msg))
	}
}
