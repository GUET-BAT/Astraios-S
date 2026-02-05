package svc

import (
	"net/http"
	"time"

	"github.com/GUET-BAT/Astraios-S/gateway-service/api/internal/config"
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
