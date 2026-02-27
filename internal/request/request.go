package request

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strings"
)

type Request struct {
	RequestLine  RequestLine
	requestState int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var VALID_METHODS = []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "CONNECT", "TRACE"}

func parseRequestLine(data []byte) (RequestLine, int, error) {

	if len(data) == 0 {
		return RequestLine{}, 0, nil
	}

	i := bytes.Index(data, []byte("\r\n"))

	if i == -1 {
		return RequestLine{}, 0, nil
	}

	requestLine := strings.Split(string(data[:i]), "\r\n")[0]
	reqParts := strings.Fields(requestLine)

	method := reqParts[0]
	path := reqParts[1]
	version := reqParts[2]
	httpVersion := strings.Split(version, "/")[1]

	if len(reqParts) != 3 {
		return RequestLine{}, 0, errors.New("invalid request line parts number")
	}

	if !slices.Contains(VALID_METHODS, method) {
		return RequestLine{}, 0, errors.New("invalid method")
	}

	if httpVersion != "1.1" {
		return RequestLine{}, 0, errors.New("invalid http version, expecting 1.1")
	}

	return RequestLine{
		Method:        method,
		RequestTarget: path,
		HttpVersion:   httpVersion,
	}, i + 2, nil

	// if len(reqParts) != 3 {
	// 	return RequestLine{}, errors.New("invalid request line parts number")
	// }

	// method := reqParts[0]
	// path := reqParts[1]
	// version := reqParts[2]

	// if !slices.Contains(VALID_METHODS, method) {
	// 	return RequestLine{}, errors.New("invalid method")
	// }

	// httpVersion := strings.Split(version, "/")[1]

	// if httpVersion != "1.1" {
	// 	return RequestLine{}, errors.New("invalid http version, expecting 1.1")
	// }

	// return RequestLine{
	// 	Method:        method,
	// 	RequestTarget: path,
	// 	HttpVersion:   httpVersion,
	// }, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	b, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	rl, _, err := parseRequestLine(b)

	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: rl,
	}, nil

}
