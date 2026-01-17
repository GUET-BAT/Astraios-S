package config

import "github.com/zeromicro/go-zero/rest"

type AuthServiceConfig struct {
	BaseUrl string
}

type Config struct {
	rest.RestConf
	AuthService AuthServiceConfig
}

