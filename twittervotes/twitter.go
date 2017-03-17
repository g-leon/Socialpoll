package main

import (
	"net"
	"time"
	"io"
)

var conn net.Conn
var reader io.ReadCloser

// dial ensures that the connection is closed and
// then opens a new connection, keeping the conn
// variable updated with the new connection.
// If a connection dies or is closed, we can
// redial without worrying about zombie connections.
func dial(netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5*time.Second)
	if err != nil {
		return nil, err
	}

	conn = netc

	return netc, nil
}

// closeConn can be called at any time in order to break
// the ongoing connection with Twitter and tidy things up.
// If the program is called with Ctrl+C then we can call
// this function just before exiting.
func closeConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

