package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/GUET-BAT/Astraios-S/user-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/user-service/pb/userpb"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	if in == nil {
		return &userpb.RegisterResponse{Code: 1}, nil
	}
	username := strings.TrimSpace(in.Username)
	password := in.Password
	if username == "" || password == "" {
		return &userpb.RegisterResponse{Code: 1}, nil
	}

	var count int64
	queryCtx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()
	err := l.svcCtx.SqlConn.QueryRowCtx(queryCtx, &count,
		`SELECT COUNT(1) FROM t_user WHERE username = ? AND deleted_at IS NULL`, username)
	if err != nil {
		l.Errorf("register query user failed: %v", err)
		return &userpb.RegisterResponse{Code: 3}, nil
	}
	if count > 0 {
		return &userpb.RegisterResponse{Code: 2}, nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := generateUserID()
	err = l.svcCtx.SqlConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if _, err := session.ExecCtx(ctx,
			`INSERT INTO t_user (id, username, password, status) VALUES (?, ?, ?, 1)`,
			userID, username, string(hashed)); err != nil {
			return err
		}
		if _, err := session.ExecCtx(ctx,
			`INSERT INTO t_user_profile (user_id) VALUES (?)`,
			userID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if isDuplicateKey(err) {
			return &userpb.RegisterResponse{Code: 2}, nil
		}
		l.Errorf("register insert user failed: %v", err)
		return &userpb.RegisterResponse{Code: 3}, nil
	}

	return &userpb.RegisterResponse{Code: 0}, nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}
