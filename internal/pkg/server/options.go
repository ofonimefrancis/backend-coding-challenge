package server

type ServerOption func(*Server)

func WithHealthChecker(hc HealthChecker) ServerOption {
	return func(s *Server) {
		s.healthChecker = hc
	}
}

func WithRouter(router RouterProvider) ServerOption {
	return func(s *Server) {
		s.router = router
	}
}
