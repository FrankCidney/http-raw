package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"strings"
)

type Handler interface {
	ServeHTTP(w *response.Writer, r *request.Request)
}

type HandlerFunc func(w *response.Writer, r *request.Request)

func (f HandlerFunc) ServeHTTP(w *response.Writer, r *request.Request) {
	f(w, r)
}

type ServeMux struct {
	handlers map[string]Handler
}

func (mux *ServeMux) ServeHTTP(w *response.Writer, r *request.Request) {
	path := r.RequestLine.RequestTarget
	// finding direct match
	if handler, ok := mux.handlers[path]; ok {
		handler.ServeHTTP(w, r)
	}

	// simple wildcard matching, handles 3 kinds of paths: 
	// 1) /path/{id} 2) /{id}/path and 3) /path/{id}/path
	var bestMatch Handler
	var bestMatchpath string
	for p, handler := range mux.handlers {
		if strings.Contains(path, p) {
			// more specific path gets priority
			if len(p) > len(bestMatchpath) {
				bestMatchpath = p
				bestMatch = handler
			}
		}
	}
	bestMatch.ServeHTTP(w, r)
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		handlers: make(map[string]Handler),
	}
}

func (mux *ServeMux) Handle(path string, handler Handler) {
	mux.handlers[path] = handler
}

func (mux *ServeMux) HandleFunc(path string, f func(w *response.Writer, r *request.Request)) {
	mux.Handle(path, HandlerFunc(f))
}
