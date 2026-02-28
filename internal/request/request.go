package request

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"tcptohttp/internal/headers"
)

type parserState string
const (
	StateInit parserState = "INIT"
	StateHeaders parserState = "HEADERS"
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
	Headers 	 *headers.Headers
	state 		 parserState
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
		Headers: headers.NewHeaders(),
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
		currentData := data[read:]

		switch r.state {
			case StateError:
				return 0, ERROR_REQUEST_IN_ERROR_STATE
			case StateHeaders:
				n, done, err := r.Headers.Parse(currentData)

				if err != nil {
					return 0, err
				}

				if n == 0 {
					break outer
				}

				read += n

				if done {
					r.state = StateDone
				}

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
		bufLen += n
		readN, parseErr := request.parse(buf[:bufLen])
		if parseErr != nil {
			return nil, parseErr
		}
		copy(buf, buf[readN:bufLen])
		bufLen -= readN

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}

	return request, nil
}
