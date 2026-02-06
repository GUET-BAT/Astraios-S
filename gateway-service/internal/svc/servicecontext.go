// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/config"
	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/middleware"
	"github.com/GUET-BAT/Astraios-S/user-service/pb/userpb"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	JwtAuth     rest.Middleware
	UserService userpb.UserServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	userClient := zrpc.MustNewClient(c.UserService)
	return &ServiceContext{
		Config:      c,
		JwtAuth:     middleware.NewJwtAuthMiddleware(c.JwtAuth).Handle,
		UserService: userpb.NewUserServiceClient(userClient.Conn()),
	}
}
