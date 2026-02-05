package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
		return &userpb.VerifyPasswordResponse{Code: 1}, fmt.Errorf("verify req is empty")
	}
	username := strings.TrimSpace(in.Username)
	password := in.Password
	if username == "" || password == "" {
		return &userpb.VerifyPasswordResponse{Code: 1}, nil
	}

	var record struct {
		ID     int64  `db:"id"`
		Hash   string `db:"password"`
		Status int    `db:"status"`
	}
	queryCtx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()
	err := l.svcCtx.SqlConn.QueryRowCtx(queryCtx, &record,
		`SELECT id, password, status FROM t_user WHERE username = ? AND deleted_at IS NULL LIMIT 1`, username)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return &userpb.VerifyPasswordResponse{Code: 1, Message: "account no exist"}, nil
		}
		return nil, err
	}
	if record.Status != 1 {
		return &userpb.VerifyPasswordResponse{Code: 1, Message: fmt.Sprintf("accout stat is %d", record.Status)}, nil
	}
	if bcrypt.CompareHashAndPassword([]byte(record.Hash), []byte(password)) != nil {
		return &userpb.VerifyPasswordResponse{Code: 1, Message: "password incorrect"}, nil
	}

	return &userpb.VerifyPasswordResponse{
		Code:   0,
		UserId: fmt.Sprintf("%d", record.ID),
		Roles:  []string{},
	}, nil
}
