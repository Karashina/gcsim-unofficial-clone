package servermode

import "sync"

// Minimal types to satisfy imports that referenced pkg/servermode.
// These are no-op placeholders after servermode removal.

type Server struct {
	sync.Mutex
	pool map[string]any
}

// isRunning returns false for all ids if pool is nil; placeholder implementation.
func (s *Server) isRunning(id string) bool {
	if s.pool == nil {
		return false
	}
	_, ok := s.pool[id]
	return ok
}

// Placeholder methods used by other files; no-op implementations.
func (s *Server) ready()    {}
func (s *Server) running()  {}
func (s *Server) validate() {}
func (s *Server) sample()   {}
func (s *Server) run()      {}
func (s *Server) latest()   {}
func (s *Server) cancel()   {}
func (s *Server) info()     {}
