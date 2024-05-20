package main

import (
	"context"
	"log"
	"sync"

	pb "therealbroker/api/proto"

	"google.golang.org/grpc"
)

const address = "localhost:8080"

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBrokerClient(conn)

	var wg sync.WaitGroup
	// Creating 5 goroutines for each of the Publish, Subscribe, and Fetch operations
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go publishMessages(&wg, c, 100) // each goroutine will publish 10000 messages
		// wg.Add(1)
		// go subscribeMessages(&wg, c, 100) // each goroutine will attempt to subscribe 10000 times
		// wg.Add(1)
		// go fetchMessages(&wg, c, 100) // each goroutine will fetch 10000 messages
	}
	wg.Wait() // Wait for all goroutines to complete
}

func publishMessages(wg *sync.WaitGroup, c pb.BrokerClient, count int) {
	defer wg.Done()
	for i := 0; i < count; i++ {
		response, err := c.Publish(context.Background(), &pb.PublishRequest{
			Subject:           "test",
			Body:              []byte("message"),
			ExpirationSeconds: int32(i),
		})
		if err != nil {
			log.Printf("Publish failed: %v", err)
		} else {
			log.Printf("Published message ID: %d", response.Id)
		}
	}
}

func subscribeMessages(wg *sync.WaitGroup, c pb.BrokerClient, count int) {
	defer wg.Done()
	for i := 0; i < count; i++ {
		stream, err := c.Subscribe(context.Background(), &pb.SubscribeRequest{Subject: "test"})
		if err != nil {
			log.Printf("Subscribe failed: %v", err)
			continue
		}
		go func() {
			msg, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				return
			}
			log.Printf("Received message: %s", string(msg.Body))
		}()
	}
}

func fetchMessages(wg *sync.WaitGroup, c pb.BrokerClient, count int) {
	defer wg.Done()
	for i := 0; i < count; i++ {
		response, err := c.Fetch(context.Background(), &pb.FetchRequest{
			Subject: "test",
			Id:      int32(i), // assuming you want to fetch by IDs incrementally
		})
		if err != nil {
			log.Printf("Fetch failed: %v", err)
			continue
		}
		log.Printf("Fetched message: %s", string(response.Body))
	}
}
