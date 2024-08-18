package app

type Server struct {
	session *Session
}

func NewServer(session *Session) *Server {
	return &Server{
		session: session,
	}
}
