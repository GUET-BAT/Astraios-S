// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"strings"

	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/middleware"
	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/types"
	"github.com/GUET-BAT/Astraios-S/user-service/pb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SetAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetAvatarLogic {
	return &SetAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetAvatarLogic) SetAvatar(req *types.AvatarUrlRequest) (resp *types.AvatarUrlResponse, err error) {
	// Step 1: Validate request parameters.
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Step 2: Read user id from token subject.
	subject, ok := middleware.SubjectFromContext(l.ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	userID := strings.TrimSpace(subject)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	// Step 3: Build RPC request to user-service.
	rpcReq := &userpb.UserAvatarRequest{
		UserId: userID,
	}

	// Step 4: Call user-service SetUserAvatar with timeout.
	ctx, cancel := context.WithTimeout(l.ctx, rpcCallTimeout)
	defer cancel()
	rpcResp, err := l.svcCtx.UserService.SetUserAvatar(ctx, rpcReq)
	if err != nil {
		l.Errorf("set avatar: rpc call failed: %v", err)
		return nil, err
	}

	// Step 5: Map RPC response to HTTP response.
	return &types.AvatarUrlResponse{
		AvatarUrl: rpcResp.AvatarUrl,
	}, nil
}
