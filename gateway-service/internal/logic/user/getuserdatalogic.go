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

type GetUserDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserDataLogic {
	return &GetUserDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserDataLogic) GetUserData(req *types.UserDataRequest) (resp *types.UserDataResponse, err error) {
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
	rpcReq := &userpb.UserDataRequest{
		UserId: userID,
	}

	// Step 4: Call user-service GetUserData with timeout.
	ctx, cancel := context.WithTimeout(l.ctx, rpcCallTimeout)
	defer cancel()
	rpcResp, err := l.svcCtx.UserService.GetUserData(ctx, rpcReq)
	if err != nil {
		l.Errorf("get user data: rpc call failed: %v", err)
		return nil, err
	}

	// Step 5: Map RPC response to HTTP response.
	return &types.UserDataResponse{
		Code: 0,
		Data: types.UserDataResponseData{
			UserInfo: types.UserInfo{
				Userid:          rpcResp.UserId,
				Nickname:        rpcResp.Nickname,
				Avatar:          rpcResp.Avatar,
				Gender:          rpcResp.Gender,
				Birthday:        rpcResp.Birthday,
				Bio:             rpcResp.Bio,
				BackgroundImage: rpcResp.BackgroundImage,
				Country:         rpcResp.Country,
				Province:        rpcResp.Province,
				City:            rpcResp.City,
				School:          rpcResp.School,
				Major:           rpcResp.Major,
				GraduationYear:  rpcResp.GraduationYear,
				CreatedAt:       rpcResp.CreatedAt,
				UpdatedAt:       rpcResp.UpdatedAt,
			},
		},
	}, nil
}
