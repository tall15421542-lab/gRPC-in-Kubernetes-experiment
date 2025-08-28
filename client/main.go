package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/deliveryhero/pd-dine-in-box/script/grpc-experiment/client/usecases"
	pb "github.com/deliveryhero/pd-dine-in-box/script/grpc-experiment/experiment"
)

var (
	addr        = flag.String("addr", "greeter-server.default.svc.cluster.local", "the address to connect to")
	port        = flag.String("port", "50051", "the port to connect to")
	isInCluster = flag.Bool("in_k8s", false, "whether the app runs in the cluster")
)

var clients = make(map[string]grpcClient)

type grpcClient interface {
	pb.GreeterClient
	GetName() string
	Close()
}

func main() {
	registerClients()
	defer closeClients()

	r := gin.Default()

	r.GET("/experiment/client", func(c *gin.Context) {
		n := c.DefaultQuery("n", "1000")
		numOfReqs, err := strconv.Atoi(n)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Invalid query string N: %s", err))
			return
		}

		clientType := c.DefaultQuery("type", "Client")
		client, ok := clients[clientType]
		if !ok {
			c.String(http.StatusBadRequest, fmt.Sprintf("client type %s is not implemented", clientType))
			return
		}

		report := runExperiment(client, numOfReqs)
		c.JSON(http.StatusOK, gin.H{
			"time":         report.time.String(),
			"count_per_ip": report.countPerIP,
		})
	})

	r.Run(":8080")
}

func registerClients() {
	flag.Parse()

	client := usecases.NewClient(*addr, *port)
	clients[client.GetName()] = client

	clientWithReuseConn := usecases.NewClientWithReuseConn(*addr, *port)
	clients[clientWithReuseConn.GetName()] = clientWithReuseConn

	if *isInCluster {
		fmt.Println("register k8s resolver client")
		clientWithKubeResolver := usecases.NewClientWithKubeResolver(*addr, *port)
		clients[clientWithKubeResolver.GetName()] = clientWithKubeResolver
	}
}

func closeClients() {
	for _, client := range clients {
		client.Close()
	}
}

type experimentReport struct {
	time       time.Duration  `json:"time"`
	countPerIP map[string]int `json:"count_per_ip"`
}

func runExperiment(client grpcClient, numOfReqs int) experimentReport {
	start := time.Now()

	countPerIP := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < numOfReqs; i = i + 1 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			r, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: client.GetName()})
			if err != nil {
				log.Printf("could not greet: %v", err)
				return
			}

			log.Printf("Greeting: %s from %s", r.GetMessage(), r.GetIpAddress())

			mu.Lock()
			countPerIP[r.GetIpAddress()] = countPerIP[r.GetIpAddress()] + 1
			mu.Unlock()
		}()
	}

	wg.Wait()

	elapsed := time.Now().Sub(start)
	fmt.Println(elapsed)

	for k, v := range countPerIP {
		fmt.Println(k, v)
	}

	return experimentReport{
		time:       elapsed,
		countPerIP: countPerIP,
	}
}
