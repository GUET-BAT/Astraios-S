// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	UserService zrpc.RpcClientConf
	AuthService zrpc.RpcClientConf
	JwtAuth     JwtAuthConf `json:",optional"`
}

type JwtAuthConf struct {
	Issuer       string `json:",optional"`
	CacheSeconds int64  `json:",default=300"`
}
