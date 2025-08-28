package usecases

import (
	"context"
	"fmt"
	"log"

	pb "github.com/deliveryhero/pd-dine-in-box/script/grpc-experiment/experiment"
	"github.com/sercand/kuberesolver/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientWithKubeResolver struct {
	conn *grpc.ClientConn
}

func NewClientWithKubeResolver(addr, port string) *ClientWithKubeResolver {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		grpc.WithBlock(),
	}

	kuberesolver.RegisterInCluster()

	target := fmt.Sprintf("kubernetes:///%s:%s", addr, port)
	conn, err := grpc.DialContext(context.Background(), target, options...)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &ClientWithKubeResolver{
		conn: conn,
	}
}

func (c *ClientWithKubeResolver) SayHello(ctx context.Context, in *pb.HelloRequest, opts ...grpc.CallOption) (*pb.HelloReply, error) {
	client := pb.NewGreeterClient(c.conn)
	r, err := client.SayHello(ctx, in)
	return r, err
}

func (c *ClientWithKubeResolver) Close() {
	c.conn.Close()
}

func (c *ClientWithKubeResolver) GetName() string {
	return "ClientWithKubeResolver"
}
