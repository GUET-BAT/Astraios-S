package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	CommonService zrpc.RpcClientConf
	ConfigDataId  string `json:",optional"`
	Sql           sqlx.SqlConf    `json:"sql"`
	Redis         redis.RedisConf `json:"redis"`
}
