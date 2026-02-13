package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GUET-BAT/Astraios-S/user-service/internal/svc"
	"github.com/GUET-BAT/Astraios-S/user-service/internal/util"
	"github.com/GUET-BAT/Astraios-S/user-service/pb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SetUserDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserDataLogic {
	return &SetUserDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetUserDataLogic) SetUserData(in *userpb.UserDataRequest) (*userpb.UserDataResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	userID := strings.TrimSpace(in.UserId)
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	parsedID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		l.Infof("set user data: invalid user_id format: %s", userID)
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	info := in.GetUserInfo()
	if info == nil {
		return nil, status.Error(codes.InvalidArgument, "user_info is required")
	}

	updates := make([]string, 0, 8)
	args := make([]any, 0, 8)

	if nickname := strings.TrimSpace(info.Nickname); nickname != "" {
		updates = append(updates, "nickname = ?")
		args = append(args, nickname)
	}
	if avatar := strings.TrimSpace(info.Avatar); avatar != "" {
		if !isHTTPURL(avatar) {
			if err := util.ValidateObjectName(avatar); err != nil {
				l.Infof("set user data: invalid avatar object key: %v", err)
				return nil, status.Error(codes.InvalidArgument, "invalid avatar object key")
			}
		}
		updates = append(updates, "avatar = ?")
		args = append(args, avatar)
	}
	if info.Gender != 0 {
		if info.Gender < 0 || info.Gender > 2 {
			return nil, status.Error(codes.InvalidArgument, "invalid gender")
		}
		updates = append(updates, "gender = ?")
		args = append(args, info.Gender)
	}
	if birthday := strings.TrimSpace(info.Birthday); birthday != "" {
		if _, err := time.Parse("2006-01-02", birthday); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid birthday format")
		}
		updates = append(updates, "birthday = ?")
		args = append(args, birthday)
	}
	if bio := strings.TrimSpace(info.Bio); bio != "" {
		updates = append(updates, "bio = ?")
		args = append(args, bio)
	}
	if backgroundImage := strings.TrimSpace(info.BackgroundImage); backgroundImage != "" {
		updates = append(updates, "background_image = ?")
		args = append(args, backgroundImage)
	}
	if country := strings.TrimSpace(info.Country); country != "" {
		updates = append(updates, "country = ?")
		args = append(args, country)
	}
	if province := strings.TrimSpace(info.Province); province != "" {
		updates = append(updates, "province = ?")
		args = append(args, province)
	}
	if city := strings.TrimSpace(info.City); city != "" {
		updates = append(updates, "city = ?")
		args = append(args, city)
	}
	if school := strings.TrimSpace(info.School); school != "" {
		updates = append(updates, "school = ?")
		args = append(args, school)
	}
	if major := strings.TrimSpace(info.Major); major != "" {
		updates = append(updates, "major = ?")
		args = append(args, major)
	}
	if info.GraduationYear != 0 {
		if info.GraduationYear < 0 {
			return nil, status.Error(codes.InvalidArgument, "invalid graduation year")
		}
		updates = append(updates, "graduation_year = ?")
		args = append(args, info.GraduationYear)
	}

	if len(updates) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no fields to update")
	}

	query := fmt.Sprintf("UPDATE t_user_profile SET %s WHERE user_id = ?", strings.Join(updates, ", "))
	execCtx, cancel := context.WithTimeout(l.ctx, dbQueryTimeout)
	defer cancel()
	if _, err := l.svcCtx.WriteConn.ExecCtx(execCtx, query, append(args, parsedID)...); err != nil {
		l.Errorf("set user data: update failed: %v", err)
		return nil, status.Error(codes.Internal, "internal error")
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
	err = l.svcCtx.WriteConn.QueryRowCtx(queryCtx, &record, `
SELECT user_id, nickname, avatar, gender, birthday, bio, background_image, country, province, city,
       school, major, graduation_year, created_at, updated_at
FROM t_user_profile
WHERE user_id = ?
LIMIT 1`, parsedID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			l.Infof("set user data: not found, userId=%s", userID)
			return nil, status.Error(codes.NotFound, "user not found")
		}
		l.Errorf("set user data: query failed: %v", err)
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
