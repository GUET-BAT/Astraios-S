// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	CommonService zrpc.RpcClientConf
	ConfigDataId  string `json:",optional"`
	UserService   zrpc.RpcClientConf
	AuthService   zrpc.RpcClientConf
	JwtAuth       JwtAuthConf     `json:",optional"`
	CacheRedis    redis.RedisConf `json:"cacheRedis,optional"`
}

type JwtAuthConf struct {
	Issuer       string `json:",optional"`
	CacheSeconds int64  `json:",default=300"`
}
