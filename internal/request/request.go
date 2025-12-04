package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"httpfromtcp/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateError   parserState = "error"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        string
	state       parserState
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

var SEPARATOR = []byte("\r\n")

const bufferSize = 8

var ErrUnsupportedHTTPVersion = fmt.Errorf("unsupported http version. use 1.1")
var ErrRequestInErrorState = fmt.Errorf("request in error state")

func newRequest() *Request {
	return &Request{
		Headers: headers.NewHeaders(),
		state:   StateInit,
	}
}

func getContentLengthInt(headers headers.Headers, name string, defaultValue int) int {
	lengthStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}

	lengthInt, err := strconv.Atoi(lengthStr)
	if err != nil {
		return defaultValue
	}
	return lengthInt
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	var err error
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))

	// verify number of parts in request line
	if len(parts) != 3 {
		err = fmt.Errorf("invalid request line: %s; want 3 parts, have %d", startLine, len(parts))
		return nil, 0, err
	}

	// verify method only contains capital alphabetic characters
	for _, char := range parts[0] {
		if char < 'A' || char > 'Z' {
			err = errors.New("invalid method")
			break
		}
	}
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %s", err, parts[0])
	}

	// verify http version (currently only supports 1.1)
	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrUnsupportedHTTPVersion
	}

	requestLine := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}
	return requestLine, read, nil
}

func (r *Request) hasBody() bool {
	length := getContentLengthInt(r.Headers, "Content-Length", 0)
	return length > 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}

		switch r.state {
		case StateError:
			return 0, ErrRequestInErrorState
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.state = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)

			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}
		case StateBody:
			length := getContentLengthInt(r.Headers, "Content-Length", 0)
			if length == 0 {
				panic("chunked not implemented")
				// r.state = StateDone
				// break outer
			}

			remaining := min(length - len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == length {
				r.state = StateDone
			}
		case StateDone:
			break outer
		default:
			panic("something went very wrong")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 0, bufferSize)
	for !request.done() {
		tempBuf := make([]byte, bufferSize)
		numRead, err := reader.Read(tempBuf)
		if err != nil {
			return nil, err
		}

		buf = append(buf, tempBuf[:numRead]...)

		numParsed, err := request.parse(buf)
		if err != nil {
			return nil, err
		}

		buf = buf[numParsed:]
	}

	return request, nil
}
