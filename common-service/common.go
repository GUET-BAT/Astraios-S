package main

import (
	"flag"
	"fmt"

	"github.com/GUET-BAT/Astraios-S/common-service/internal/config"
	"github.com/GUET-BAT/Astraios-S/common-service/internal/server"
	"github.com/GUET-BAT/Astraios-S/common-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/common-service/pb/commonpb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/common.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		commonpb.RegisterCommonServiceServer(grpcServer, server.NewCommonServiceServer(ctx))

		// go-zero's zrpc.MustNewServer already registers the gRPC health service internally.
		// Do NOT register grpc_health_v1.HealthServer again here, or it will panic with
		// "duplicate service registration for grpc.health.v1.Health".

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
