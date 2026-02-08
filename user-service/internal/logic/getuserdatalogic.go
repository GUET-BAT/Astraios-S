package logic

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/GUET-BAT/Astraios-S/user-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/user-service/pb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetUserDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserDataLogic {
	return &GetUserDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserDataLogic) GetUserData(in *userpb.UserDataRequest) (*userpb.UserDataResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	// NOTE: Field name will change to in.UserId after proto regeneration.
	userID := strings.TrimSpace(in.Userid)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	parsedID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		l.Infof("get user data: invalid user_id format: %s", userID)
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	var record struct {
		UserID          int64          `db:"user_id"`
		Nickname        sql.NullString `db:"nickname"`
		Avatar          sql.NullString `db:"avatar"`
		Gender          sql.NullInt64  `db:"gender"`
		Birthday        sql.NullTime   `db:"birthday"`
		Bio             sql.NullString `db:"bio"`
		BackgroundImage sql.NullString `db:"background_image"`
		Country         sql.NullString `db:"country"`
		Province        sql.NullString `db:"province"`
		City            sql.NullString `db:"city"`
		School          sql.NullString `db:"school"`
		Major           sql.NullString `db:"major"`
		GraduationYear  sql.NullInt64  `db:"graduation_year"`
		CreatedAt       sql.NullTime   `db:"created_at"`
		UpdatedAt       sql.NullTime   `db:"updated_at"`
	}

	queryCtx, cancel := context.WithTimeout(l.ctx, dbQueryTimeout)
	defer cancel()
	err = l.svcCtx.ReadConn.QueryRowCtx(queryCtx, &record, `
SELECT user_id, nickname, avatar, gender, birthday, bio, background_image, country, province, city,
       school, major, graduation_year, created_at, updated_at
FROM t_user_profile
WHERE user_id = ?
LIMIT 1`, parsedID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			l.Infof("get user data: not found, userId=%s", userID)
			return nil, status.Error(codes.NotFound, "user not found")
		}
		l.Errorf("get user data: query failed: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &userpb.UserDataResponse{
		UserId:          strconv.FormatInt(record.UserID, 10),
		Nickname:        nullString(record.Nickname),
		Avatar:          nullString(record.Avatar),
		Gender:          nullInt32(record.Gender),
		Birthday:        formatDate(record.Birthday),
		Bio:             nullString(record.Bio),
		BackgroundImage: nullString(record.BackgroundImage),
		Country:         nullString(record.Country),
		Province:        nullString(record.Province),
		City:            nullString(record.City),
		School:          nullString(record.School),
		Major:           nullString(record.Major),
		GraduationYear:  nullInt32(record.GraduationYear),
		CreatedAt:       formatTime(record.CreatedAt),
		UpdatedAt:       formatTime(record.UpdatedAt),
	}, nil
}

func nullString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func nullInt32(value sql.NullInt64) int32 {
	if value.Valid {
		return int32(value.Int64)
	}
	return 0
}

func formatDate(value sql.NullTime) string {
	if value.Valid {
		return value.Time.Format("2006-01-02")
	}
	return ""
}

func formatTime(value sql.NullTime) string {
	if value.Valid {
		return value.Time.Format(time.RFC3339)
	}
	return ""
}
