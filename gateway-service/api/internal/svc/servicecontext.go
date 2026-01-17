package svc

import (
	"net/http"
	"time"

	"github.com/astraio/astraios-gateway/api/internal/config"
)

type ServiceContext struct {
	Config     config.Config
	HttpClient *http.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	return &ServiceContext{
		Config:     c,
		HttpClient: client,
	}
}
