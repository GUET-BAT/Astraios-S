package logic

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/GUET-BAT/Astraios-S/user-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/user-service/pb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SetUserAvatarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserAvatarLogic {
	return &SetUserAvatarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetUserAvatarLogic) SetUserAvatar(in *userpb.UserAvatarRequest) (*userpb.UserAvatarResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	userID := strings.TrimSpace(in.UserId)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	parsedID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		l.Infof("set user avatar: invalid user_id format: %s", userID)
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	var exists int64
	queryCtx, cancel := context.WithTimeout(l.ctx, dbQueryTimeout)
	defer cancel()
	err = l.svcCtx.ReadConn.QueryRowCtx(queryCtx, &exists, `
SELECT 1
FROM t_user_profile
WHERE user_id = ?
LIMIT 1`, parsedID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			l.Infof("set user avatar: not found, userId=%s", userID)
			return nil, status.Error(codes.NotFound, "user not found")
		}
		l.Errorf("set user avatar: query failed: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	objectKey, err := buildAvatarObjectKey(userID)
	if err != nil {
		l.Errorf("set user avatar: generate object key failed: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	if l.svcCtx.OSSClient == nil {
		l.Errorf("set user avatar: oss client not initialized")
		return nil, status.Error(codes.Internal, "internal error")
	}
	ossCtx, cancel := context.WithTimeout(l.ctx, ossOpTimeout)
	defer cancel()
	presign, err := l.svcCtx.OSSClient.PresignPut(ossCtx, objectKey, avatarUploadExpiry)
	if err != nil {
		l.Errorf("set user avatar: presign put failed: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &userpb.UserAvatarResponse{AvatarUrl: presign.URL}, nil
}
