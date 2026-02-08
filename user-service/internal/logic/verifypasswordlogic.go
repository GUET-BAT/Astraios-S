package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/GUET-BAT/Astraios-S/user-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/user-service/pb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
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

func (l *VerifyPasswordLogic) VerifyPassword(in *userpb.VerifyPasswordRequest) (*userpb.VerifyPasswordResponse, error) {
	if in == nil {
		return nil, fmt.Errorf("request is required")
	}
	username := strings.TrimSpace(in.Username)
	password := in.Password
	if username == "" || password == "" {
		return &userpb.VerifyPasswordResponse{Code: CodeInvalidParam, Message: "invalid credentials"}, nil
	}

	var record struct {
		ID     int64  `db:"id"`
		Hash   string `db:"password"`
		Status int    `db:"status"`
	}
	queryCtx, cancel := context.WithTimeout(l.ctx, dbQueryTimeout)
	defer cancel()
	err := l.svcCtx.ReadConn.QueryRowCtx(queryCtx, &record,
		`SELECT id, password, status FROM t_user WHERE username = ? AND deleted_at IS NULL LIMIT 1`, username)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			l.Infof("verify password: account not found, username=%s", username)
			return &userpb.VerifyPasswordResponse{Code: CodeInvalidParam, Message: "invalid credentials"}, nil
		}
		l.Errorf("verify password: query failed: %v", err)
		return nil, err
	}
	if record.Status != 1 {
		l.Infof("verify password: account disabled, username=%s status=%d", username, record.Status)
		return &userpb.VerifyPasswordResponse{Code: CodeInvalidParam, Message: "invalid credentials"}, nil
	}
	if bcrypt.CompareHashAndPassword([]byte(record.Hash), []byte(password)) != nil {
		l.Infof("verify password: incorrect password, username=%s", username)
		return &userpb.VerifyPasswordResponse{Code: CodeInvalidParam, Message: "invalid credentials"}, nil
	}

	l.Infof("verify password: success, username=%s userId=%d", username, record.ID)
	return &userpb.VerifyPasswordResponse{
		Code:   CodeSuccess,
		UserId: fmt.Sprintf("%d", record.ID),
		Roles:  []string{},
	}, nil
}
