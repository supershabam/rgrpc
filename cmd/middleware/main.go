package main

import (
	"net"

	"golang.org/x/net/context"

	"github.com/baddeploy/unixgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server is used to implement reader.Reader
type server struct{}

// ReadAll implements reader.ReadAll
func (s *server) Test(ctx context.Context, in *unixgrpc.TestRequest) (*unixgrpc.TestReply, error) {
	return &unixgrpc.TestReply{
		Greeting: "HI",
	}, nil
}

func main() {
	l, err := net.Listen("unix", "/tmp/echo.sock")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	unixgrpc.RegisterTesterServer(s, &server{})
	reflection.Register(s)
	err = s.Serve(l)
	if err != nil {
		panic(err)
	}
}
