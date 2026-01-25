package network

import (
	"bufio"
	"fmt"
	"mmorpg/internal/game"
	"net"
	"sync"
)

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	wg         sync.WaitGroup
	game       *game.Game
}

func NewServer(listenAddr string, g *game.Game) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		game:       g,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	fmt.Printf("Server running on %s\n", s.listenAddr)

	go s.acceptLoop()

	<-s.quitch
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-s.quitch:
				return
			default:
				fmt.Printf("accept error: %s\n", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	fmt.Printf("new connection from %s\n", conn.RemoteAddr())

	player := s.game.AddPlayer(conn)
	defer s.game.RemovePlayer(player.ID)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		HandleCommand(player, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("connection error: %v\n", err)
	}
}

func (s *Server) Stop() {
	close(s.quitch)
	s.ln.Close()
	s.wg.Wait()
}
