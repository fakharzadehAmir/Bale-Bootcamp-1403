package server

import (
	"therealbroker/api/proto"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

func NewBrokerServer(brokerServer proto.BrokerServer) *grpc.Server {
	grpcMetrics := grpc_prometheus.NewServerMetrics()
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcMetrics.UnaryServerInterceptor()),
		grpc.StreamInterceptor(grpcMetrics.StreamServerInterceptor()),
	)
	grpc_prometheus.Register(grpcServer)
	proto.RegisterBrokerServer(grpcServer, brokerServer)
	grpcMetrics.InitializeMetrics(grpcServer)
	return grpcServer
}
