package main

import (
	"net"
	"time"
	"io"
)

var conn net.Conn
func dial(netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5 * time.Second)
	if err != nil {
		return nil, err
	}
	conn = netc
	return netc, nil
}

// Not Only closing connection, but also closing reader which reads response
var reader io.ReadCloser
func closeConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}
