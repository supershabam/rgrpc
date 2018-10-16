package main

import (
	"context"
	"fmt"
	"net"

	"github.com/supershabam/rgrpc"
	"github.com/supershabam/rgrpc/examples/hello"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const addr = "localhost:9001"

func main() {
	err := run(context.Background())
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("while running client")
	}
}

func run(ctx context.Context) error {
	l, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(l)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	d, err := rgrpc.NewDialer(lis)
	if err != nil {
		return err
	}
	// treat this like a listener.Accept loop
	for {
		cc, err := grpc.DialContext(ctx, "",
			grpc.WithDialer(d.Dial),
			// MUST instruct grpc to allow insecure. The underlaying
			// connection may be secure, but grpc won't know that.
			grpc.WithInsecure(),
			// MUST instruct grpc to block before returning
			// a client connection or else this for loop will
			// become an infinite for loop of compute resource
			// theft.
			grpc.WithBlock(),
		)
		if err != nil {
			return err
		}
		go handle(ctx, cc)
	}
}

func handle(ctx context.Context, cc *grpc.ClientConn) error {
	defer cc.Close()
	client := hello.NewHelloClient(cc)
	resp, err := client.Greet(ctx, &hello.Person{
		Name: "Oliver Twist",
	})
	if err != nil {
		return err
	}
	fmt.Printf("received greeting: %s\n", resp.Phase)
	return nil
}
