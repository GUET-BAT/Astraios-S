package svc

import (
	"sync"

	"common-service/internal/config"
	"common-service/internal/nacos"
)

type ServiceContext struct {
	Config config.Config

	nacosOnce   sync.Once
	nacosClient *nacos.Client
	nacosErr    error
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}

func (s *ServiceContext) NacosClient() (*nacos.Client, error) {
	s.nacosOnce.Do(func() {
		s.nacosClient, s.nacosErr = nacos.NewClientFromEnv()
	})
	return s.nacosClient, s.nacosErr
}
