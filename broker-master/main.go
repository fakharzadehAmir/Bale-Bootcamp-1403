package main

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"therealbroker/api/server"
	"therealbroker/config"
	"therealbroker/pkg/database"
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
	config.SetConfigInstance(cfg)
}
func main() {
	ctx := context.Background()

	log.Infof("requested storage type is %s\n", cfg.Broker.StorageType)

	//	Prometheus created metrics initialization
	go middleware.EvaluateEnvMetrics()
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(cfg.Prometheus.Port), nil)
		if err != nil {
			log.WithError(err).Fatalf("prometheus can not listen on the given port %v\n", cfg.Prometheus.Port)
		}
	}()

	//	Jager created metrics initialization
	closer, err := middleware.NewJaegerObject(*config.GetConfigInstance(), log)
	if err != nil {
		log.WithError(err).Fatalln("can not create a tracer object")
	}
	log.Infoln("jaeger tracer object created successfully")
	defer closer.Close()
	opentracing.SetGlobalTracer(middleware.Tracer)

	//	Initial PostgresDB
	dbInstance, err := database.ConnectToPg(ctx, config.GetConfigInstance(), log)
	if err != nil {
		log.WithError(err).Fatalln("could not connect to the postgres")
	}
	log.Infof("connected to database successfully on port %v\n", cfg.PostgresDB.Port)
	defer func() {
		if err := dbInstance.Close(); err != nil {
			log.WithError(err).Warn("Failed to close postgres database connection")
		}
	}()

	//	Initial CassandraDB
	cassandraDbInstance, err := database.ConnectToCassandra(log)
	if err != nil {
		log.WithError(err).Fatalln("could not connect to the cassandra")
	}
	log.Infof("connected to cassandra database successfully on port %v\n", cfg.CassandraDB.Port)
	defer func() {
		cassandraDbInstance.Close()
	}()

	//	Initial Broker Module
	brokerServer := server.NewImplementedServer()
	log.Infoln("broker server object created successfully")

	//	Initialize RPC APIs
	grpcServer := server.NewBrokerServer(brokerServer)
	log.Infoln("broker grpc server created successfully")

	// Set up a listener for the gRPC server
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Broker.Port))
	if err != nil {
		log.WithError(err).Fatalf("Failed to listen on port %v\n", cfg.Broker.Port)
	}
	log.Infof("gRPC server is listening on port %v\n", cfg.Broker.Port)

	// Serve gRPC Server
	if err := grpcServer.Serve(listener); err != nil {
		log.WithError(err).Fatalf("Failed to start gRPC server")
	}

	// Graceful shutdown handling
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// <-c
	// log.Println("Shutting down gRPC server...")
	// grpcServer.GracefulStop()
	// log.Println("Server successfully stopped")
}
