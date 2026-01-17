package handler

import (
	"net/http"

	"github.com/astraio/astraios-gateway/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, ctx *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/v1/web/profile",
			Handler: NewAuthMiddleware(ctx)(WebProfileHandler(ctx)),
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/v1/mobile/profile",
			Handler: NewAuthMiddleware(ctx)(MobileProfileHandler(ctx)),
		},
	})
}
