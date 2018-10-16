package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/supershabam/rgrpc"
	"github.com/supershabam/rgrpc/examples/hello"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const target = "localhost:9001"

// the dialer acts like a gprc server but is actually initiating network
// connectivity to the listener.
func main() {
	err := run(context.Background())
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("while running example server")
	}
}

func run(ctx context.Context) error {
	l, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(l)
	d := dialer()
	lis, err := rgrpc.NewListener(d)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	hello.RegisterHelloServer(grpcServer, &server{})
	go func() {
		<-ctx.Done()
		grpcServer.Stop()
	}()
	return grpcServer.Serve(lis)
}

type server struct{}

func (s *server) Greet(ctx context.Context, p *hello.Person) (*hello.Greeting, error) {
	return &hello.Greeting{
		Phase: fmt.Sprintf("greetings %s", p.Name),
	}, nil
}

// dialer is how we establish a connection to our target "client." If
// the dialer ever returns an error, the server will stop serving.
// So, we want to continuously retry instead of actually returning
// an error.
// After the connection closes, the dialer will be called again immediately
// which we don't want to cause flooding of dials against the target so
// we must also rate limit successes.
func dialer() func(ctx context.Context) (net.Conn, error) {
	d := &net.Dialer{}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(0) // infinite retry
	return func(ctx context.Context) (net.Conn, error) {
		var conn net.Conn
		op := func() error {
			zap.L().With(
				zap.String("target", target),
			).Debug("dialing")
			c, err := d.DialContext(ctx, "tcp", target)
			if err != nil {
				return err
			}
			conn = c
			return nil
		}
		// TODO backoff.Retry immediately executes op even if the backoff says to delay
		err := backoff.RetryNotify(
			op,
			backoff.WithContext(&noResetBackOff{b}, ctx),
			func(err error, d time.Duration) {
				zap.L().With(
					zap.Error(err),
					zap.Duration("duration", d),
					zap.String("target", target),
				).Debug("retryable error while dialing")
			})
		if err != nil {
			return nil, err
		}
		// TODO scheduler b.Reset() after some time in successful state
		return conn, nil
	}
}

type noResetBackOff struct {
	BackOff backoff.BackOff
}

// NextBackOff returns the duration to wait before retrying the operation,
// or backoff. Stop to indicate that no more retries should be made.
//
// Example usage:
//
// 	duration := backoff.NextBackOff();
// 	if (duration == backoff.Stop) {
// 		// Do not retry operation.
// 	} else {
// 		// Sleep for duration and retry operation.
// 	}
//
func (b *noResetBackOff) NextBackOff() time.Duration {
	return b.BackOff.NextBackOff()
}

// Reset to initial state.
func (b *noResetBackOff) Reset() {
	return
}
