// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"strings"
	"time"

	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/middleware"
	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	blacklistMinTTL     = time.Minute
	blacklistDefaultTTL = 24 * time.Hour
	redisOpTimeout      = 2 * time.Second
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.LogoutRequest) (resp *types.LogoutResponse, err error) {
	// Step 1: Validate request parameters.
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if l.svcCtx.Redis == nil {
		l.Errorf("logout: redis client not configured")
		return nil, status.Error(codes.Internal, "internal error")
	}

	// Step 2: Read token from context (set by JwtAuth middleware).
	token, ok := middleware.TokenFromContext(l.ctx)
	if !ok || strings.TrimSpace(token) == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	// Step 3: Determine blacklist TTL from token expiration.
	ttl := blacklistDefaultTTL
	if exp, ok := middleware.TokenExpiryFromContext(l.ctx); ok {
		ttl = time.Until(exp)
	}
	if ttl <= 0 {
		ttl = blacklistMinTTL
	}
	seconds := int(ttl.Seconds())
	if seconds < 1 {
		seconds = 1
	}

	// Step 4: Store token in blacklist.
	ctx, cancel := context.WithTimeout(l.ctx, redisOpTimeout)
	defer cancel()
	if err := l.svcCtx.Redis.SetexCtx(ctx, middleware.TokenBlacklistKey(token), "1", seconds); err != nil {
		l.Errorf("logout: set token blacklist failed: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	// Step 5: Return success response.
	return &types.LogoutResponse{
		Code: 0,
		Msg:  "ok",
	}, nil
}
