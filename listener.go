// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package rgrpc

import (
	"context"
	"net"
	"sync"
)

// Listener implements net.Listener but it doesn't listen in the traditional sense.
// Listener actively dials a target with the provided dialer to establish a
// net.Conn. When the conn is closed, a the dialer is called again to establish
// a new net.Conn.
type Listener struct {
	dialer func(context.Context) (net.Conn, error)

	cancel context.CancelFunc
	connch <-chan net.Conn
	ctx    context.Context
	wait   func() error
	m      sync.Mutex
}

func NewListener(dialer func(context.Context) (net.Conn, error)) (*Listener, error) {
	return &Listener{
		dialer: dialer,
	}, nil
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	l.init()
	c, active := <-l.connch
	if !active {
		return nil, l.wait()
	}
	return c, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *Listener) Close() error {
	l.init()
	l.cancel()
	for c := range l.connch {
		c.Close()
	}
	return l.wait()
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return &raddr{}
}

func (l *Listener) init() {
	l.m.Lock()
	defer l.m.Unlock()
	if l.ctx != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	connch, wait := l.conns(ctx)
	l.ctx = ctx
	l.cancel = cancel
	l.connch = connch
	l.wait = wait
}

func (l *Listener) conns(ctx context.Context) (<-chan net.Conn, func() error) {
	// normally I would use an errgroup here, but I don't want to introduce a dependency to run this library
	errch := make(chan error, 1)
	onErr := func(err error) {
		select {
		case errch <- err:
		default:
		}
	}
	out := make(chan net.Conn)
	go func() {
		defer close(out)
		defer close(errch)
		for {
			// it is the job of the caller to provide a Dialer that implements exponential
			// backoff
			c, err := l.dialer(ctx)
			if err != nil {
				onErr(err)
				return
			}
			// wrap the conn as a cconn so we can monitor when it closes
			cc := newCcon(c)
			// send the conn to the consumer
			select {
			case <-ctx.Done():
				cc.Close() // we own the connection right now so we have to close it if the context closes
				onErr(ctx.Err())
				return
			case out <- cc:
			}
			// wait for the connection to close before looping
			select {
			case <-ctx.Done():
				onErr(ctx.Err())
				return
			case <-cc.Done():
			}
		}
	}()
	wait := func() error {
		select {
		case err, active := <-errch:
			if !active {
				return nil
			}
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	// memoize the result of wait and ensure it's only called once so that the caller may call
	// it many times
	var monce sync.Once
	mready := make(chan struct{})
	var merr error
	mwait := func() error {
		monce.Do(func() {
			merr = wait()
			close(mready)
		})
		<-mready
		return merr
	}
	return out, mwait
}
