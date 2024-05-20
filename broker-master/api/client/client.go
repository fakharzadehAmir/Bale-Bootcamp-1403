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
	address     = "localhost:8080"
	workerCount = 100
	jobCount    = 60000
	runDuration = 2 * time.Minute // Run the worker pool for 2 minutes
)

type Job struct {
	jobType string
	id      int
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBrokerClient(conn)

	var wg sync.WaitGroup
	jobQueue := make(chan Job, jobCount)

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(&wg, jobQueue, c)
	}

	// Add jobs to the queue
	for i := 0; i < jobCount; i++ {
		jobQueue <- Job{jobType: "publish", id: i}
		// jobQueue <- Job{jobType: "subscribe", id: i}
		// jobQueue <- Job{jobType: "fetch", id: i}
	}

	// Run the worker pool for the specified duration
	ticker := time.NewTicker(runDuration)
	go func() {
		<-ticker.C
		close(jobQueue) // Signal to workers to stop processing after the duration
	}()

	wg.Wait()
}

func worker(wg *sync.WaitGroup, jobs <-chan Job, c pb.BrokerClient) {
	defer wg.Done()
	for job := range jobs {
		switch job.jobType {
		case "publish":
			publishMessage(c, job.id)
		case "subscribe":
			subscribeMessage(c, job.id)
		case "fetch":
			fetchMessage(c, job.id)
		}
	}
}

func publishMessage(c pb.BrokerClient, id int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	response, err := c.Publish(ctx, &pb.PublishRequest{
		Subject:           "test",
		Body:              []byte("message"),
		ExpirationSeconds: int32(id),
	})
	if err != nil {
		log.Printf("Publish failed: %v", err)
	} else {
		log.Printf("Published message ID: %d", response.Id)
	}
}

func subscribeMessage(c pb.BrokerClient, id int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stream, err := c.Subscribe(ctx, &pb.SubscribeRequest{Subject: "test"})
	if err != nil {
		log.Printf("Subscribe failed: %v", err)
		return
	}
	msg, err := stream.Recv()
	if err != nil {
		log.Printf("Error receiving message: %v", err)
		return
	}
	log.Printf("Received message: %s", string(msg.Body))
}

func fetchMessage(c pb.BrokerClient, id int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	response, err := c.Fetch(ctx, &pb.FetchRequest{
		Subject: "test",
		Id:      int32(id),
	})
	if err != nil {
		log.Printf("Fetch failed: %v", err)
		return
	}
	log.Printf("Fetched message: %s", string(response.Body))
}
