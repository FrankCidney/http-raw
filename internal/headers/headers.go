package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

var SEPARATOR = []byte("\r\n")
var ErrInvalidFieldLine = fmt.Errorf("malformed field line")

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Get(name string) (string, bool) {
	name = strings.ToLower(name)
	str, ok := h[name]
	return str, ok
}

func (h Headers) Replace(name, value string) {
	name = strings.ToLower(name)
	h[name] = value
}

func (h Headers) Delete(name string) {
	name = strings.ToLower(name)
	delete(h, name)
}

func (h Headers) Set(name, value string) {
	name = strings.ToLower(name)
	if v, ok := h[strings.ToLower(name)]; ok {
		value = fmt.Sprintf("%s, %s", v, value)
	}

	h[name] = value
}

func (h Headers) ForEach(f func(n, v string)) {
	for n, v := range h {
		f(n, v)
	}
}

func isToken(str []byte) bool {
	found := false
	for _, ch := range str {
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {
			found = true
		}

		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}
	return true
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", ErrInvalidFieldLine
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) || bytes.HasSuffix(name, []byte("\t")) {
		return "", "", fmt.Errorf("malformed field name")
	}

	return string(name), string(value), nil
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	for {
		idx := bytes.Index(data[n:], SEPARATOR)
		if idx == -1 {
			break
		}

		if idx == 0 {
			n += len(SEPARATOR)
			done = true
			break
		}

		name, value, err := parseHeader(data[n:idx+n])
		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("malformed field name")
		}

		n += idx + len(SEPARATOR)
		h.Set(name, value)
	}

	// fmt.Println("n:",n)
	// fmt.Println("done:", done)
	return n, done, nil
}	
