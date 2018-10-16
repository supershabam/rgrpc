package rgrpc

import (
	"context"
	"net"
	"sync"
	"time"
)

// Dialer is actually a traditional listening server but returns inbound connections
// as if you dialed them whenever you call Dial. It is expected that the caller
// call Dial frequently to handle multiple connections just like a server would
// traditionally call Accept multiple times.
type Dialer struct {
	cancel context.CancelFunc
	connch <-chan net.Conn
	ctx    context.Context
	err    error
	lis    net.Listener
	m      sync.Mutex
}

func NewDialer(lis net.Listener) (*Dialer, error) {
	return &Dialer{
		lis: lis,
	}, nil
}

func (d *Dialer) Dial(_ string, dur time.Duration) (net.Conn, error) {
	d.init()
	tctx := context.Background()
	if dur != 0 {
		c, cancel := context.WithTimeout(context.Background(), dur)
		defer cancel()
		tctx = c
	}
	select {
	case <-d.ctx.Done():
		return nil, d.err
	case <-tctx.Done():
		// TODO make temporary
		return nil, tctx.Err()
	case conn := <-d.connch:
		return conn, nil
	}
}

func (d *Dialer) init() {
	d.m.Lock()
	defer d.m.Unlock()
	if d.ctx != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	connch := make(chan net.Conn)
	go func() {
		for {
			conn, err := d.lis.Accept()
			if err != nil {
				d.err = err
				cancel()
				return
			}
			select {
			case connch <- conn:
			}
		}
	}()
	d.ctx = ctx
	d.cancel = cancel
	d.connch = connch
}
