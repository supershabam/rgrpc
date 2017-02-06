package main

import (
	"errors"
	"net"
	"sync/atomic"

	"golang.org/x/net/context"

	"sync"

	"github.com/baddeploy/unixgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server is used to implement reader.Reader
type server struct {
	conns map[uint64]net.Conn
	l     sync.RWMutex
	sid   uint64
}

func newServer() *server {
	return &server{
		conns: map[uint64]net.Conn{},
	}
}

// ReadAll implements reader.ReadAll
func (s *server) Test(ctx context.Context, in *unixgrpc.TestRequest) (*unixgrpc.TestReply, error) {
	return &unixgrpc.TestReply{
		Greeting: "HI",
	}, nil
}

func (s *server) Dial(ctx context.Context, in *unixgrpc.DialRequest) (*unixgrpc.DialReply, error) {
	conn, err := net.Dial(in.Network, in.Address)
	if err != nil {
		return nil, err
	}
	sid := atomic.AddUint64(&s.sid, 1)
	s.l.Lock()
	s.conns[sid] = conn
	s.l.Unlock()
	return &unixgrpc.DialReply{
		Sid: sid,
	}, nil
}

func (s *server) Read(ctx context.Context, in *unixgrpc.ReadRequest) (*unixgrpc.ReadReply, error) {
	s.l.RLock()
	conn, ok := s.conns[in.Sid]
	s.l.RUnlock()
	if !ok {
		return nil, errors.New("not found")
	}
	deadline, ok := ctx.Deadline()
	if ok {
		conn.SetReadDeadline(deadline)
	}
	buf := make([]byte, in.N)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return &unixgrpc.ReadReply{
		Buf: buf[:n],
	}, nil
}

func (s *server) Write(ctx context.Context, in *unixgrpc.WriteRequest) (*unixgrpc.WriteReply, error) {
	s.l.RLock()
	conn, ok := s.conns[in.Sid]
	s.l.RUnlock()
	if !ok {
		return nil, errors.New("not found")
	}
	deadline, ok := ctx.Deadline()
	if ok {
		conn.SetWriteDeadline(deadline)
	}
	n, err := conn.Write(in.Buf)
	if err != nil {
		return nil, err
	}
	return &unixgrpc.WriteReply{
		N: int32(n),
	}, nil
}

func (s *server) Close(ctx context.Context, in *unixgrpc.CloseRequest) (*unixgrpc.CloseReply, error) {
	s.l.RLock()
	conn, ok := s.conns[in.Sid]
	s.l.RUnlock()
	if !ok {
		return nil, errors.New("not found")
	}
	return &unixgrpc.CloseReply{}, conn.Close()
}

func main() {
	l, err := net.Listen("unix", "/tmp/echo.sock")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	unixgrpc.RegisterTesterServer(s, newServer())
	reflection.Register(s)
	err = s.Serve(l)
	if err != nil {
		panic(err)
	}
}
