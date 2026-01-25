package network

import (
	"net"
	"sync"
)

// ExportedHandleConnection exposes handleConnection for testing.
func (s *Server) ExportedHandleConnection(conn net.Conn) {
	s.handleConnection(conn)
}

// GetWG returns the waitgroup for testing.
func (s *Server) GetWG() *sync.WaitGroup {
	return &s.wg
}
