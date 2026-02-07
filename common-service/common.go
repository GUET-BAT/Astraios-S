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
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
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

		// gRPC 健康检查服务，供 K8s 的 gRPC 探针调用。
		// 这里标记为 SERVING，即表示服务就绪。
		healthServer := health.NewServer()
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
