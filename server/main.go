package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/deliveryhero/pd-dine-in-box/script/grpc-experiment/experiment"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	time.Sleep(500 * time.Millisecond)
	return &pb.HelloReply{
		Message:   "Hello " + in.GetName(),
		IpAddress: os.Getenv("MY_POD_IP"),
	}, nil
}

func main() {
	ipAddr, ok := os.LookupEnv("MY_POD_IP")
	if !ok {
		log.Fatal("Env [MY_POD_IP] does not exist")
	}

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %s:%v", ipAddr, *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
