package webserver

import (
	"net"

	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func NewGRPCServer() (*grpc.Server, net.Listener, error) {
	listener, err := net.Listen(viper.GetString("rpc_server_network"),
		viper.GetString("rpc_server_host")+":"+viper.GetString("rpc_server_port"))
	if err != nil {
		return nil, nil, err
	}

	rpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	return rpcServer, listener, nil
}
