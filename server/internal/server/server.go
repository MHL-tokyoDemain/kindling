package server

type Server struct {
	port int
}

func New(port int) *Server {
	return &Server{port: port}
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) Shutdown() error {
	return nil
}
