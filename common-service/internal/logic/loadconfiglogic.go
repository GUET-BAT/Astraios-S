package logic

import (
	"context"
	"strings"
	"time"

	"github.com/GUET-BAT/Astraios-S/common-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/common-service/pb/commonpb"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	codeSuccess        int32 = 0
	codeFailed         int32 = 1
	defaultRequestTimeout    = 5 * time.Second
)

type LoadConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoadConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoadConfigLogic {
	return &LoadConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoadConfigLogic) LoadConfig(in *commonpb.LoadConfigRequest) (*commonpb.LoadConfigResponse, error) {
	if in == nil || strings.TrimSpace(in.NacosDataId) == "" {
		return &commonpb.LoadConfigResponse{
			Code:    codeFailed,
			Message: "nacosDataId is required",
		}, nil
	}

	client, err := l.svcCtx.NacosClient()
	if err != nil {
		l.Errorf("load config: init nacos client failed: %v", err)
		return &commonpb.LoadConfigResponse{
			Code:    codeFailed,
			Message: "nacos client initialization failed",
		}, nil
	}

	ctx, cancel := context.WithTimeout(l.ctx, defaultRequestTimeout)
	defer cancel()

	cfg, err := client.LoadConfig(ctx, in.NacosDataId)
	if err != nil {
		l.Errorf("load config: fetch %s failed: %v", in.NacosDataId, err)
		return &commonpb.LoadConfigResponse{
			Code:    codeFailed,
			Message: "failed to load config from nacos",
		}, nil
	}

	return &commonpb.LoadConfigResponse{
		Code:    codeSuccess,
		Message: "ok",
		Config:  cfg,
	}, nil
}
