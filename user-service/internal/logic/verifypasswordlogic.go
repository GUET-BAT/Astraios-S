package logic

import (
	"context"

	"user-service/internal/svc"
	"user-service/pb/github.com/astraios/grpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyPasswordLogic {
	return &VerifyPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VerifyPasswordLogic) VerifyPassword(in *user.VerifyPasswordRequest) (*user.VerifyPasswordResponse, error) {
	// todo: add your logic here and delete this line

	return &user.VerifyPasswordResponse{}, nil
}
