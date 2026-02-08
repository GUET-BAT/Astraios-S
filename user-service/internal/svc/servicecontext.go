package svc

import (
	"github.com/GUET-BAT/Astraios-S/user-service/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config  config.Config
	SqlConn sqlx.SqlConn
	Redis   *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		SqlConn: sqlx.MustNewConn(c.Sql),
		Redis:   redis.MustNewRedis(c.CacheRedis),
	}
}
