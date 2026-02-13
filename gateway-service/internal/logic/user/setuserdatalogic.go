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

type SetUserDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetUserDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserDataLogic {
	return &SetUserDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetUserDataLogic) SetUserData(req *types.UserDataRequest) (resp *types.UserDataResponse, err error) {
	// Step 1: Validate request parameters.
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	userID := strings.TrimSpace(req.Userid)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "userid is required")
	}
	if !hasUserInfo(req.UserInfo) {
		return nil, status.Error(codes.InvalidArgument, "user_info is required")
	}

	// Step 2: Enforce that token subject matches requested user id.
	subject, ok := middleware.SubjectFromContext(l.ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	if subject != userID {
		l.Infof("set user data: subject mismatch, subject=%s requested=%s", subject, userID)
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}

	// Step 3: Build RPC request to user-service.
	rpcReq := &userpb.UserDataRequest{
		UserId: userID,
		UserInfo: &userpb.UserInfo{
			Nickname:        strings.TrimSpace(req.UserInfo.Nickname),
			Avatar:          strings.TrimSpace(req.UserInfo.Avatar),
			Gender:          req.UserInfo.Gender,
			Birthday:        strings.TrimSpace(req.UserInfo.Birthday),
			Bio:             strings.TrimSpace(req.UserInfo.Bio),
			BackgroundImage: strings.TrimSpace(req.UserInfo.BackgroundImage),
			Country:         strings.TrimSpace(req.UserInfo.Country),
			Province:        strings.TrimSpace(req.UserInfo.Province),
			City:            strings.TrimSpace(req.UserInfo.City),
			School:          strings.TrimSpace(req.UserInfo.School),
			Major:           strings.TrimSpace(req.UserInfo.Major),
			GraduationYear:  req.UserInfo.GraduationYear,
		},
	}

	// Step 4: Call user-service SetUserData with timeout.
	ctx, cancel := context.WithTimeout(l.ctx, rpcCallTimeout)
	defer cancel()
	rpcResp, err := l.svcCtx.UserService.SetUserData(ctx, rpcReq)
	if err != nil {
		l.Errorf("set user data: rpc call failed: %v", err)
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

func hasUserInfo(info types.UserInfo) bool {
	checks := []func() bool{
		func() bool { return strings.TrimSpace(info.Nickname) != "" },
		func() bool { return strings.TrimSpace(info.Avatar) != "" },
		func() bool { return strings.TrimSpace(info.Birthday) != "" },
		func() bool { return strings.TrimSpace(info.Bio) != "" },
		func() bool { return strings.TrimSpace(info.BackgroundImage) != "" },
		func() bool { return strings.TrimSpace(info.Country) != "" },
		func() bool { return strings.TrimSpace(info.Province) != "" },
		func() bool { return strings.TrimSpace(info.City) != "" },
		func() bool { return strings.TrimSpace(info.School) != "" },
		func() bool { return strings.TrimSpace(info.Major) != "" },

		func() bool { return info.Gender != 0 },
		func() bool { return info.GraduationYear != 0 },
	}

	for _, check := range checks {
		if check() {
			return true
		}
	}

	return false
}
