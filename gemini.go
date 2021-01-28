package main

import (
	"crypto/tls"
	"io"
	"io/ioutil"
)

var config = tls.Config{
	InsecureSkipVerify: true,
}

func isdigit(b byte) bool {
	return '0' <= b && b <= '9'
}

// make a request and return the response code
func request(host, req string) (int, error) {
        conn, err := tls.Dial("tcp", host, &config)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	if _, err := io.WriteString(conn, req); err != nil {
		return 0, err
	}

	res := make([]byte, 2)
	if _, err := conn.Read(res); err != nil {
		return 0, err
	}

	code := 0
	if isdigit(res[0]) && isdigit(res[1]) {
		code = int((res[0] - '0') * 10 + (res[1] - '0'))
	}
	return code, drain(conn)
}

func drain(src io.Reader) error {
	_, err := io.Copy(ioutil.Discard, src)
	return err
}
