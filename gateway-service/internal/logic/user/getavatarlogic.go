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

type GetAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAvatarLogic {
	return &GetAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAvatarLogic) GetAvatar(req *types.AvatarUrlRequest) (resp *types.AvatarUrlResponse, err error) {
	// Step 1: Validate request parameters.
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	userID := strings.TrimSpace(req.Userid)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "userid is required")
	}

	// Step 2: Enforce that token subject matches requested user id.
	subject, ok := middleware.SubjectFromContext(l.ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	if subject != userID {
		l.Infof("get avatar: subject mismatch, subject=%s requested=%s", subject, userID)
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}

	// Step 3: Build RPC request to user-service.
	rpcReq := &userpb.UserAvatarRequest{
		UserId: userID,
	}

	// Step 4: Call user-service GetUserAvatar with timeout.
	ctx, cancel := context.WithTimeout(l.ctx, rpcCallTimeout)
	defer cancel()
	rpcResp, err := l.svcCtx.UserService.GetUserAvatar(ctx, rpcReq)
	if err != nil {
		l.Errorf("get avatar: rpc call failed: %v", err)
		return nil, err
	}

	// Step 5: Map RPC response to HTTP response.
	return &types.AvatarUrlResponse{
		AvatarUrl: rpcResp.AvatarUrl,
	}, nil
}
