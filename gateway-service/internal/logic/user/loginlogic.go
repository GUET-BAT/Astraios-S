// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/types"
	"github.com/GUET-BAT/Astraios-S/gateway-service/pb/authpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	if req == nil || req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "invaid params")
	}

	rpcReq := &authpb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	ctx, cancel := context.WithTimeout(l.ctx, rpcCallTimeout)
	defer cancel()

	rpcResp, err := l.svcCtx.AuthService.Login(ctx, rpcReq)
	if err != nil {
		l.Logger.Error("Login rpc request failed, err: %v", err)
		return nil, err
	}

	return &types.LoginResponse{
		Code: 0,
		Data: types.LoginResponseData{
			AccessToken:  rpcResp.AccessToken,
			RefreshToken: rpcResp.RefreshToken}}, err
}
