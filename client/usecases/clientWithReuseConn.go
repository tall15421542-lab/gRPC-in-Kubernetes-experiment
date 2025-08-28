package usecases

import (
	"context"
	"log"

	pb "github.com/deliveryhero/pd-dine-in-box/script/grpc-experiment/experiment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientWithReuseConn struct {
	conn *grpc.ClientConn
}

func NewClientWithReuseConn(addr, port string) *ClientWithReuseConn {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	target := addr + ":" + port
	conn, err := grpc.DialContext(context.Background(), target, options...)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &ClientWithReuseConn{
		conn: conn,
	}
}

func (c *ClientWithReuseConn) SayHello(ctx context.Context, in *pb.HelloRequest, opts ...grpc.CallOption) (*pb.HelloReply, error) {
	client := pb.NewGreeterClient(c.conn)
	r, err := client.SayHello(ctx, in)
	return r, err
}

func (c *ClientWithReuseConn) Close() {
	c.conn.Close()
}

func (c *ClientWithReuseConn) GetName() string {
	return "ClientWithReuseConn"
}
