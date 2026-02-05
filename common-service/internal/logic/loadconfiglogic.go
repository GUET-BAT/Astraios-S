package logic

import (
	"context"
	"strings"
	"time"

	"github.com/GUET-BAT/Astraios-S/common-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/common-service/pb/commonpb"

	"github.com/zeromicro/go-zero/core/logx"
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
			Code:    1,
			Message: "nacosDataId is required",
		}, nil
	}

	client, err := l.svcCtx.NacosClient()
	if err != nil {
		l.Errorf("load nacos config: %v", err)
		return &commonpb.LoadConfigResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	ctx, cancel := context.WithTimeout(l.ctx, defaultRequestTimeout)
	defer cancel()

	cfg, err := client.LoadConfig(ctx, in.NacosDataId)
	if err != nil {
		l.Errorf("load nacos config for %s: %v", in.NacosDataId, err)
		return &commonpb.LoadConfigResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &commonpb.LoadConfigResponse{
		Code:    0,
		Message: "ok",
		Config:  cfg,
	}, nil
}

const defaultRequestTimeout = 5 * time.Second
