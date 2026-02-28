package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

var rn = []byte("\r\n")


func NewHeaders() *Headers {
	return &Headers {
		headers: map[string]string{},
	}
}

func isToken(str []byte) bool {
	for _, b := range str {
		found := false
		if b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9'|| bytes.ContainsAny([]byte("!#$%&'*+-.^_`|~"), string(b)) {
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
		return "", "", fmt.Errorf("invalid field line: %s", string(fieldLine))
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field line: %s", string(fieldLine))
	}

	return string(name), string(value), nil	
}

func (h *Headers) Get(name string) (string) {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)

	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[name] = value
	}
}

func (h *Headers) ForEach(f func(name, value string)) {
	for name, value := range h.headers {
		f(name, value)
	}
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break;
		}

		if idx == 0 {
			done = true
			break;
		}

		name, value, err := parseHeader(data[read:read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("invalid header name: %s", name)
		}

		read += idx + len(rn)
		h.Set(name, value)

	}
	return read, done, nil

}