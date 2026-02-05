package main

import (
	"flag"

	"github.com/GUET-BAT/Astraios-S/gateway-service/api/internal/config"
	"github.com/GUET-BAT/Astraios-S/gateway-service/api/internal/handler"
	"github.com/GUET-BAT/Astraios-S/gateway-service/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	server.Start()
}
