package main

import (
	"context"
	"log"
	"sync"
	"time"

	pb "therealbroker/api/proto"

	"google.golang.org/grpc"
)

const (
	address        = "localhost:8080"
	workerCount    = 1000
	jobCount       = 50000
	messagesPerJob = 1
	timeout        = 2 * time.Minute
	targetRate     = 50000.0           // Target rate in requests per minute
	targetRPS      = targetRate / 60.0 // Target rate in requests per second
)

type Job struct {
	jobType string
	count   int
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewBrokerClient(conn)

	jobQueue := make(chan Job, jobCount)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(&wg, jobQueue, client)
	}

	ticker := time.NewTicker(time.Duration(1 * time.Millisecond))
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			jobQueue <- Job{jobType: "publish", count: messagesPerJob}
		}
	}()

	time.Sleep(timeout)
	close(jobQueue)
	wg.Wait()
}

func worker(wg *sync.WaitGroup, jobs <-chan Job, client pb.BrokerClient) {
	defer wg.Done()
	for job := range jobs {
		switch job.jobType {
		case "publish":
			publishMessages(client, job.count)
		case "subscribe":
			subscribeMessages(client, job.count)
		case "fetch":
			fetchMessages(client, job.count)
		}
	}
}

func publishMessages(client pb.BrokerClient, count int) {
	for i := 0; i < count; i++ {
		response, err := client.Publish(context.Background(), &pb.PublishRequest{
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

func subscribeMessages(client pb.BrokerClient, count int) {
	for i := 0; i < count; i++ {
		stream, err := client.Subscribe(context.Background(), &pb.SubscribeRequest{Subject: "test"})
		if err != nil {
			log.Printf("Subscribe failed: %v", err)
			continue
		}
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return
		}
		log.Printf("Received message: %s", string(msg.Body))
	}
}

func fetchMessages(client pb.BrokerClient, count int) {
	for i := 0; i < count; i++ {
		response, err := client.Fetch(context.Background(), &pb.FetchRequest{
			Subject: "test",
			Id:      int32(i),
		})
		if err != nil {
			log.Printf("Fetch failed: %v", err)
			continue
		}
		log.Printf("Fetched message: %s", string(response.Body))
	}
}
