package usecases

import (
	"context"
	"fmt"

	pb "github.com/deliveryhero/pd-dine-in-box/script/grpc-experiment/experiment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Address string
	Options []grpc.DialOption
}

func NewClient(addr, port string) *Client {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	return &Client{
		Address: addr + ":" + port,
		Options: options,
	}
}

func (c *Client) SayHello(ctx context.Context, in *pb.HelloRequest, opts ...grpc.CallOption) (*pb.HelloReply, error) {
	conn, err := grpc.DialContext(ctx, c.Address, c.Options...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial grpc connection to %s: %w", c.Address, err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)
	r, err := client.SayHello(ctx, in)
	return r, err
}

func (c *Client) Close() {
	fmt.Println("noop")
}

func (c *Client) GetName() string {
	return "Client"
}
