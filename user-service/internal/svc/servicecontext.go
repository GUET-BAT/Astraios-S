package svc

import (
	"fmt"

	"github.com/GUET-BAT/Astraios-S/user-service/internal/config"
	"github.com/GUET-BAT/Astraios-S/user-service/internal/util"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	ReadConn  sqlx.SqlConn // Read-only connection (read port)
	WriteConn sqlx.SqlConn // Read-write connection (write port)
	Redis     *redis.Redis
	OSSClient *util.OSSClient
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	ossClient, err := mustNewOssClient(c.Oss)
	if err != nil {
		return nil, err
	}

	return &ServiceContext{
		Config:    c,
		ReadConn:  mustNewSQLConn(c.Mysql, c.Mysql.RPort),
		WriteConn: mustNewSQLConn(c.Mysql, c.Mysql.WPort),
		Redis:     mustNewRedisClient(c.CacheRedis),
		OSSClient: ossClient,
	}, nil
}

func mustNewSQLConn(config config.MysqlConf, port int) sqlx.SqlConn {
	return sqlx.NewMysql(buildDSN(config, port))
}

func mustNewRedisClient(config redis.RedisConf) *redis.Redis {
	return redis.MustNewRedis(config)
}

func mustNewOssClient(config config.OssConf) (*util.OSSClient, error) {
	ossClient, ossErr := util.NewOSSClient(util.OSSConfig{
		Region:     config.Region,
		BucketName: config.Bucketname,
		BucketURL:  config.Bucketurl,
		Endpoint:   config.Endpoint,
	})
	if ossErr != nil {
		return nil, fmt.Errorf("failed to initialize OSS client: %w", ossErr)
	}
	return ossClient, nil
}

// buildDSN constructs a MySQL DSN string from MysqlConf with the given port.
func buildDSN(m config.MysqlConf, port int) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		m.Username, m.Password, m.Address, port, m.Database, m.Params)
}
