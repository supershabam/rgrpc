package rgrpc

import (
	"errors"
	"net"
)

// Listener implements net.Listener but it doesn't listen in the traditional sense.
// Listener actively dials a target with the provided dialer to establish a
// net.Conn. When the conn is closed, a the dialer is called again to establish
// a new net.Conn.
type Listener struct{}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	return nil, errors.New("not implemented")
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *Listener) Close() error {
	return errors.New("not implemented")
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return &raddr{}
}
