package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	CommonService zrpc.RpcClientConf
	ConfigDataId  string          `json:",optional"`
	Mysql         MysqlConf       `json:"mysql,optional"`
	CacheRedis    redis.RedisConf `json:"cacheRedis,optional"`
	Oss           OssConf         `json:"oss,optional"`
}

// MysqlConf holds read/write split MySQL configuration.
// Address is the MySQL host, WPort is the write port, RPort is the read port.
type MysqlConf struct {
	Address  string `json:"address"`
	WPort    int    `json:"wport"`
	RPort    int    `json:"rport"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Params   string `json:"params"`
}

type OssConf struct {
	Region     string `json:"region"`
	Bucketname string `json:"bucketname"`
	Bucketurl  string `json:"bucketurl"`
	Endpoint   string `json:"endpoint"`
}
