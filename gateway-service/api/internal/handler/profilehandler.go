package handler

import (
	"encoding/json"
	"net/http"

	"github.com/astraio/astraios-gateway/api/internal/svc"
	"github.com/astraio/astraios-gateway/api/internal/types"
)

func WebProfileHandler(ctx *svc.ServiceContext) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		resp := types.UserProfileResponse{
			Id:     user.Id,
			Name:   user.Name,
			Client: "web",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func MobileProfileHandler(ctx *svc.ServiceContext) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		resp := types.UserProfileResponse{
			Id:     user.Id,
			Name:   user.Name,
			Client: "mobile",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
