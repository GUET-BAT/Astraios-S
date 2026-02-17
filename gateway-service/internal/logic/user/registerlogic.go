// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"time"

	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/types"
	httpstatuscode "github.com/GUET-BAT/Astraios-S/global/http"
	"github.com/GUET-BAT/Astraios-S/user-service/pb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
)

const rpcCallTimeout = 5 * time.Second

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	// Step 1: Validate request payload.
	if req == nil || req.Username == "" || req.Password == "" {
		return &types.RegisterResponse{Code: httpstatuscode.CodeInvalidParam}, nil
	}

	// Step 2: Build RPC request to user-service.
	rpcReq := &userpb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Type:     req.Type,
	}

	// Step 3: Call user-service Register with timeout.
	ctx, cancel := context.WithTimeout(l.ctx, rpcCallTimeout)
	defer cancel()
	rpcResp, err := l.svcCtx.UserService.Register(ctx, rpcReq)
	if err != nil {
		l.Errorf("register: rpc call failed: %v", err)
		return &types.RegisterResponse{Code: httpstatuscode.CodeInternalError}, nil
	}

	// Step 4: Map RPC response to HTTP response.
	var msg string
	switch rpcResp.Code {
	case 0:
		msg = "注册成功"
	case 1:
		msg = "参数无效"
	case 2:
		msg = "用户名已存在"
	case 3:
		msg = "内部错误"
	default:
		msg = "注册失败"
	}
	return &types.RegisterResponse{
		Code: rpcResp.Code,
		Msg:  msg}, nil
}
