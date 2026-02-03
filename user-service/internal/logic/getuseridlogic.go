package logic

import (
	"context"

	"user-service/internal/svc"
	"user-service/pb/github.com/astraios/grpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserIdLogic {
	return &GetUserIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserIdLogic) GetUserId(in *user.UserDataRequest) (*user.UserDataResponse, error) {
	// todo: add your logic here and delete this line

	return &user.UserDataResponse{}, nil
}
