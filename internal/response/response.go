package response

import (
	"fmt"
	"io"
	"strconv"

	"httpfromtcp/internal/headers"
)

type StatusCode int

type Writer struct{
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

var (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// func (w *Writer) Write(b []byte) (int, error) {
// 	w.data = append(w.data, b...)
// 	return len(b), nil
// }

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var statusLine []byte

	switch statusCode {
	case StatusOk:
		statusLine = ([]byte("HTTP/1.1 200 OK\r\n"))
	case StatusBadRequest:
		statusLine = ([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case StatusInternalServerError:
		statusLine = ([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		return fmt.Errorf("unknown status code")
	}
	_, err := w.writer.Write(statusLine)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	contentLenStr := strconv.Itoa(contentLen)
	h.Set("Content-Length", contentLenStr)
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	b := []byte{}
	
	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})
	b = fmt.Append(b, "\r\n")

	_, err := w.writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	return n, err
}

// func (w *Writer) WriteTrailers(h headers.Headers) error {
// 	err := w.WriteHeaders(h)
// 	return err
// }

// func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
// 	n, err := w.writer.Write([]byte(fmt.Sprintf("%x\r\n%v\r\n", len(p), p)))
// 	// n, err = w.writer.Write(p)
// 	// n, err = w.writer.Write([]byte("\r\n"))
// 	return n, err
// }

// func (w *Writer) WriteChunkedBodyDone() (int, error) {
// 	n, err := w.writer.Write([]byte(fmt.Sprintf("%x\r\n\r\n", 0)))
// 	return n, err
// }
