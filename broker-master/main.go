package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"therealbroker/api/client"
	"therealbroker/api/server"
	"therealbroker/config"
	"therealbroker/pkg/middleware"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// Main requirements:
//  1. All tests should be passed
//  2. Your logs should be accessible in Graylog
//  3. Basic prometheus metrics ( latency, throughput, etc. ) should be implemented
//     for every base functionality ( publish, subscribe etc. )
var (
	cfg = &config.Config{}
	log = logrus.New()
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err.Error())
	}
	if err = cleanenv.ReadEnv(cfg); err != nil {
		panic(err)
	}
}

func main() {
	// ctx := context.Background()
	//	Graylog

	//	Prometheus created metrics initialization
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(cfg.Prometheus.Port), nil)
		if err != nil {
			log.WithError(err).Fatalf("prometheus can not listen on the given port %v\n", cfg.Prometheus.Port)
		}
	}()

	//	Jager created metrics initialization
	closer, err := middleware.NewJaegerObject(*cfg, log)
	if err != nil {
		log.WithError(err).Fatalln("can not create a tracer object")
	}
	log.Infoln("jaeger tracer object created successfully")
	defer closer.Close()
	opentracing.SetGlobalTracer(middleware.Tracer)

	//	Initial DB

	//	Initial CassandraDB

	//	Initial Broker Module
	brokerServer := server.NewImplementedServer()
	log.Infoln("broker grpc server created successfully")

	//	Initialize RPC APIs
	grpcServer := client.NewBrokerClient(brokerServer)
	log.Infoln("broker grpc client created successfully")

	// Set up a listener for the gRPC server
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Broker.Port))
	if err != nil {
		log.WithError(err).Fatalf("Failed to listen on port %v\n", cfg.Broker.Port)
	}
	log.Infof("gRPC server is listening on port %v\n", cfg.Broker.Port)

	// Serve gRPC Server
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.WithError(err).Fatalf("Failed to start gRPC server")
		}
	}()

	// Graceful shutdown handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("Server successfully stopped")
}
