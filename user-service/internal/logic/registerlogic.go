package logic

import (
	"context"
	"errors"
	"strings"

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
		return &userpb.RegisterResponse{Code: CodeInvalidParam}, nil
	}
	username := strings.TrimSpace(in.Username)
	password := in.Password
	if username == "" || password == "" {
		return &userpb.RegisterResponse{Code: CodeInvalidParam}, nil
	}
	if err := validateUsername(username); err != nil {
		l.Infof("register: invalid username: %v", err)
		return &userpb.RegisterResponse{Code: CodeInvalidParam}, nil
	}
	if err := validatePassword(password); err != nil {
		l.Infof("register: invalid password: %v", err)
		return &userpb.RegisterResponse{Code: CodeInvalidParam}, nil
	}

	var count int64
	queryCtx, cancel := context.WithTimeout(l.ctx, dbQueryTimeout)
	defer cancel()
	// Use WriteConn for existence check to avoid false negatives from replication lag.
	err := l.svcCtx.WriteConn.QueryRowCtx(queryCtx, &count,
		`SELECT COUNT(1) FROM t_user WHERE username = ? AND deleted_at IS NULL`, username)
	if err != nil {
		l.Errorf("register: query user failed: %v", err)
		return &userpb.RegisterResponse{Code: CodeInternal}, nil
	}
	if count > 0 {
		return &userpb.RegisterResponse{Code: CodeAlreadyExists}, nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		l.Errorf("register: hash password failed: %v", err)
		return nil, err
	}

	userID := generateUserID()
	insertCtx, insertCancel := context.WithTimeout(context.Background(), dbQueryTimeout)
	defer insertCancel()
	err = l.svcCtx.WriteConn.TransactCtx(insertCtx, func(ctx context.Context, session sqlx.Session) error {
		if _, err := session.ExecCtx(ctx,
			`INSERT INTO t_user (id, username, password, status) VALUES (?, ?, ?, 1)`,
			userID, username, string(hashed)); err != nil {
			return err
		}
		if _, err := session.ExecCtx(ctx,
			`INSERT INTO t_user_profile (user_id, avatar) VALUES (?, ?)`,
			userID, "avatars/default_avatar.jpg"); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if isDuplicateKey(err) {
			return &userpb.RegisterResponse{Code: CodeAlreadyExists}, nil
		}
		l.Errorf("register: insert user failed: %v", err)
		return &userpb.RegisterResponse{Code: CodeInternal}, nil
	}

	l.Infof("register: success, username=%s userId=%d", username, userID)
	return &userpb.RegisterResponse{Code: CodeSuccess}, nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}
