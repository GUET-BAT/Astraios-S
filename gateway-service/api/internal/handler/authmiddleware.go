package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/astraio/astraios-gateway/api/internal/svc"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

type AuthUser struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewAuthMiddleware(ctx *svc.ServiceContext) func(HandlerFunc) http.HandlerFunc {
	return func(next HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			token := parts[1]
			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, ctx.Config.AuthService.BaseUrl+"/auth/validate", nil)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := ctx.HttpClient.Do(req)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			var user AuthUser
			err = json.NewDecoder(resp.Body).Decode(&user)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctxReq := r.WithContext(NewUserContext(r.Context(), user))
			next(w, ctxReq)
		}
	}
}
