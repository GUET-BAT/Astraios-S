package svc

import (
	"fmt"

	"github.com/GUET-BAT/Astraios-S/user-service/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	ReadConn  sqlx.SqlConn // Read-only connection (read port)
	WriteConn sqlx.SqlConn // Read-write connection (write port)
	Redis     *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		ReadConn:  sqlx.NewMysql(buildDSN(c.Mysql, c.Mysql.RPort)),
		WriteConn: sqlx.NewMysql(buildDSN(c.Mysql, c.Mysql.WPort)),
		Redis:     redis.MustNewRedis(c.CacheRedis),
	}
}

// buildDSN constructs a MySQL DSN string from MysqlConf with the given port.
func buildDSN(m config.MysqlConf, port int) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		m.Username, m.Password, m.Address, port, m.Database, m.Params)
}
