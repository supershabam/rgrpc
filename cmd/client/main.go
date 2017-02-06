package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/baddeploy/unixgrpc"

	"io"

	"google.golang.org/grpc"
)

func main() {
	// Set up a connection to the server.
	dial := func(addr string, t time.Duration) (net.Conn, error) {
		return net.Dial("unix", "/tmp/echo.sock")
	}
	client, err := grpc.Dial("", grpc.WithInsecure(), grpc.WithDialer(dial))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer client.Close()
	c := unixgrpc.NewTesterClient(client)
	conn, err := unixgrpc.NewConn(&unixgrpc.Config{
		Address:      "localhost:9003",
		Network:      "tcp",
		TesterClient: c,
		Timeout:      time.Minute,
	})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, conn)
}
