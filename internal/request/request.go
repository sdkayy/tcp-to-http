package request

import (
	"bytes"
	"errors"
	"io"
	"slices"
)

type parserState string
const (
	StateInit parserState = "INIT"
	StateDone parserState = "DONE"
	StateError parserState = "ERROR"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine  RequestLine
	state parserState
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
	}
}

var VALID_METHODS = []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "CONNECT", "TRACE"}
var SEPERATOR = []byte("\r\n")

var ERROR_INVALID_REQUEST_LINE = errors.New("invalid request line parts number")
var ERROR_INVALID_DATA_SIZE = errors.New("invalid data size, expecting more data")
var ERROR_INVALID_METHOD = errors.New("invalid method")
var ERROR_INVALID_HTTP_VERSION = errors.New("invalid http version, expecting 1.1")
var ERROR_REQUEST_IN_ERROR_STATE = errors.New("in error state")

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, SEPERATOR)

	if idx == -1 {
		return nil, 0, nil
	}

	startLine := data[:idx]
	read := idx + len(SEPERATOR)

	parts := bytes.Split(startLine, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, ERROR_INVALID_REQUEST_LINE
	}

	method := string(parts[0])
	if !slices.Contains(VALID_METHODS, method) {
		return nil, 0, ERROR_INVALID_METHOD
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_INVALID_REQUEST_LINE
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {

	read := 0

outer:
	for {
		switch r.state {
			case StateError:
				return 0, ERROR_REQUEST_IN_ERROR_STATE
			case StateInit:
				rl, n, err := parseRequestLine(data[read:])
				if err != nil {
					r.state = StateError
					return 0, err
				}

				if n == 0 {
					break outer
				}

				r.RequestLine = *rl
				read += n

				r.state = StateDone

			case StateDone:
				break outer
		}
	}

	return read, nil

}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest();
	// NOTE: buffer could get overrun
	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() { 
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n;
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
