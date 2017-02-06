package unixgrpc

import (
	"errors"
	"net"
	"time"

	"golang.org/x/net/context"
)

type Config struct {
	Address      string
	Network      string
	TesterClient TesterClient
	Timeout      time.Duration
}

func NewConn(cfg *Config) (*Conn, error) {
	ctx := context.Background()
	tctx, _ := context.WithTimeout(context.Background(), cfg.Timeout)
	r, err := cfg.TesterClient.Dial(tctx, &DialRequest{
		Address: cfg.Address,
		Network: cfg.Network,
	})
	if err != nil {
		return nil, err
	}
	return &Conn{
		ctx:     ctx,
		sid:     r.Sid,
		tc:      cfg.TesterClient,
		timeout: cfg.Timeout,
	}, nil
}

type Conn struct {
	ctx     context.Context
	sid     uint64
	tc      TesterClient
	timeout time.Duration
}

type pipeAddr int

func (pipeAddr) Network() string {
	return "pipe"
}

func (pipeAddr) String() string {
	return "pipe"
}

// Read reads data from the connection.
// Read can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(b []byte) (n int, err error) {
	r, err := c.tc.Read(context.Background(), &ReadRequest{
		Sid: c.sid,
		N:   int32(len(b)),
	})
	if err != nil {
		return -1, err
	}
	return copy(b, r.Buf), nil
}

// Write writes data to the connection.
// Write can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *Conn) Write(b []byte) (n int, err error) {
	r, err := c.tc.Write(context.Background(), &WriteRequest{
		Sid: c.sid,
		Buf: b,
	})
	if err != nil {
		return -1, err
	}
	return int(r.N), nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *Conn) Close() error {
	_, err := c.tc.Close(context.Background(), &CloseRequest{
		Sid: c.sid,
	})
	return err
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	var pa pipeAddr
	return pa
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	var pa pipeAddr
	return pa
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future I/O, not just
// the immediately following call to Read or Write.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *Conn) SetDeadline(t time.Time) error {
	return errors.New("not implemented")
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return errors.New("not implemented")
}

// SetWriteDeadline sets the deadline for future Write calls.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return errors.New("not implemented")
}
