package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func toString(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

func respond400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func respond200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}



func handleYourProblem(w *response.Writer, r *request.Request) {
	h := response.GetDefaultHeaders(0)
	body := respond400()

	h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
	h.Replace("Content-Type", "text/html")
	w.WriteStatusLine(response.StatusBadRequest)
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handleMyProblem(w *response.Writer, r *request.Request) {
	h := response.GetDefaultHeaders(0)
	body := respond500()

	h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
	h.Replace("Content-Type", "text/html")
	w.WriteStatusLine(response.StatusInternalServerError)
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handleVideo(w *response.Writer, r *request.Request) {
	h := response.GetDefaultHeaders(0)
	path := filepath.Join("assets", "vim.mp4")
	data, err := os.ReadFile(path)
	if err != nil {
		handleMyProblem(w, r)
	} else {
		h.Replace("Content-Type", "video/mp4")
		h.Replace("Content-Length", fmt.Sprintf("%d", len(data)))
		w.WriteStatusLine(response.StatusOk)
		w.WriteHeaders(h)
		w.WriteBody(data)
	}
}

func handleHttpbinStream(w *response.Writer, r *request.Request) {
	h := response.GetDefaultHeaders(0)
	target := r.RequestLine.RequestTarget
	res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
	if err != nil {
		handleMyProblem(w, r)
	} else {
		w.WriteStatusLine(response.StatusOk)
		h.Delete("Content-Length")
		h.Set("Transfer-Encoding", "chunked")
		h.Replace("Content-Type", "text/plain")
		h.Set("Trailer", "X-Content-SHA256")
		h.Set("Trailer", "X-Content-Length")
		w.WriteHeaders(h)

		fullBody := []byte{}
		for {
			data := make([]byte, 32)
			n, err := res.Body.Read(data)
			if err != nil {
				break
			}
			fullBody = append(fullBody, data[:n]...)
			w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
			w.WriteBody(data[:n])
			w.WriteBody([]byte("\r\n"))
		}
		w.WriteBody([]byte("0\r\n\r\n"))
		trailers := headers.NewHeaders()
		sum := sha256.Sum256(fullBody)
		trailers.Set("X-Content-SHA256", toString(sum[:]))
		trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
		w.WriteTrailers(trailers)
	}
}

func main() {
	mux := server.NewServeMux()

	mux.HandleFunc("/yourproblem", handleYourProblem)
	mux.HandleFunc("/myproblem", handleMyProblem)
	mux.HandleFunc("/video", handleVideo)
	mux.HandleFunc("/httpbin/stream", handleHttpbinStream)

	server, err := server.Serve(port, mux)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
