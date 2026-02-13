package logic

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/GUET-BAT/Astraios-S/user-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/user-service/internal/util"
	"github.com/GUET-BAT/Astraios-S/user-service/pb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetUserAvatarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserAvatarLogic {
	return &GetUserAvatarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserAvatarLogic) GetUserAvatar(in *userpb.UserAvatarRequest) (*userpb.UserAvatarResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	userID := strings.TrimSpace(in.UserId)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	parsedID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		l.Infof("get user avatar: invalid user_id format: %s", userID)
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	var avatar sql.NullString
	queryCtx, cancel := context.WithTimeout(l.ctx, dbQueryTimeout)
	defer cancel()
	err = l.svcCtx.ReadConn.QueryRowCtx(queryCtx, &avatar, `
SELECT avatar
FROM t_user_profile
WHERE user_id = ?
LIMIT 1`, parsedID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			l.Infof("get user avatar: not found, userId=%s", userID)
			return nil, status.Error(codes.NotFound, "user not found")
		}
		l.Errorf("get user avatar: query failed: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	avatarValue := strings.TrimSpace(avatar.String)
	if !avatar.Valid || avatarValue == "" {
		return &userpb.UserAvatarResponse{AvatarUrl: ""}, nil
	}
	if isHTTPURL(avatarValue) {
		return &userpb.UserAvatarResponse{AvatarUrl: avatarValue}, nil
	}
	if err := util.ValidateObjectName(avatarValue); err != nil {
		l.Errorf("get user avatar: invalid avatar object key: %v", err)
		return nil, status.Error(codes.Internal, "invalid avatar object key")
	}

	if l.svcCtx.OSSClient == nil {
		l.Errorf("get user avatar: oss client not initialized")
		return nil, status.Error(codes.Internal, "internal error")
	}
	ossCtx, cancel := context.WithTimeout(l.ctx, ossOpTimeout)
	defer cancel()
	presign, err := l.svcCtx.OSSClient.PresignGet(ossCtx, avatarValue, avatarDisplayExpiry)
	if err != nil {
		l.Errorf("get user avatar: presign get failed: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &userpb.UserAvatarResponse{AvatarUrl: presign.URL}, nil
}
