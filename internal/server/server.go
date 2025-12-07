package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

// type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	listener  net.Listener
	handler   Handler
	hasClosed atomic.Bool
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func newServer(handler Handler) *Server {
	server := &Server{
		handler: handler,
	}
	server.hasClosed.Store(false)
	return server
}

// func writeHandlerError(w io.Writer, handlerErr HandlerError) {
// 	err := fmt.Sprintf("%d %s", handlerErr.statusCode, handlerErr.message)
// 	w.Write([]byte(err))
// }

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	headers := response.GetDefaultHeaders(0)
	req, err := request.RequestFromReader(conn)
	fmt.Println("request:", req)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(headers)
		return
	}

	s.handler.ServeHTTP(responseWriter, req)
}

func (s *Server) listen() {
	for {
		if s.hasClosed.Load() {
			return
		}

		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		go s.handle(conn)
	}
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := newServer(handler)
	server.listener = ln
	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	err := s.listener.Close()
	if err == nil {
		s.hasClosed.Store(true)
	}
	return err
}
