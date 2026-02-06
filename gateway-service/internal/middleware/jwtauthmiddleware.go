// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type JwtAuthMiddleware struct {
	cfg       config.JwtAuthConf
	mu        sync.RWMutex
	jwks      jwk.Set
	fetchedAt time.Time
}

type ctxKey int

const (
	ctxKeySubject ctxKey = iota
)

func NewJwtAuthMiddleware(cfg config.JwtAuthConf) *JwtAuthMiddleware {
	return &JwtAuthMiddleware{
		cfg: cfg,
	}
}

func (m *JwtAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Parse token from Authorization header.
		tokenStr, err := bearerToken(r)
		if err != nil {
			writeUnauthorized(w, err)
			return
		}

		// Step 2: Load JWK set (cached with TTL).
		jwks, err := m.getJwks(r.Context())
		if err != nil {
			writeUnauthorized(w, err)
			return
		}

		// Step 3: Parse and validate token signature + claims.
		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, errors.New("invalid jwt signing method")
			}
			kid, _ := token.Header["kid"].(string)
			if kid == "" {
				return nil, errors.New("missing kid in jwt header")
			}
			keys, ok := jwks.LookupKeyID(kid)
			if !ok || len(keys) == 0 {
				return nil, errors.New("unknown kid")
			}
			var key any
			if err := keys[0].Raw(&key); err != nil {
				return nil, err
			}
			return key, nil
		}, jwt.WithIssuer(m.cfg.Issuer))
		if err != nil {
			writeUnauthorized(w, err)
			return
		}

		subject, err := claims.GetSubject()
		if err != nil || strings.TrimSpace(subject) == "" {
			writeUnauthorized(w, errors.New("missing subject in jwt claims"))
			return
		}

		// Step 4: Enforce access token type when present.
		if tokenType, ok := claims["token_type"].(string); ok && tokenType != "access" {
			writeUnauthorized(w, errors.New("invalid token type"))
			return
		}

		// Step 5: Continue request.
		ctx := context.WithValue(r.Context(), ctxKeySubject, subject)
		next(w, r.WithContext(ctx))
	}
}

func writeUnauthorized(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnauthorized)
	httpx.WriteJson(w, map[string]string{"message": err.Error()})
}

func (m *JwtAuthMiddleware) getJwks(ctx context.Context) (jwk.Set, error) {
	m.mu.RLock()
	if m.jwks != nil && !m.isExpired() {
		defer m.mu.RUnlock()
		return m.jwks, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.jwks != nil && !m.isExpired() {
		return m.jwks, nil
	}

	if m.cfg.JwksUrl == "" {
		return nil, errors.New("jwks url is required")
	}

	set, err := jwk.Fetch(ctx, m.cfg.JwksUrl)
	if err != nil {
		return nil, err
	}
	m.jwks = set
	m.fetchedAt = time.Now()
	return m.jwks, nil
}

func (m *JwtAuthMiddleware) isExpired() bool {
	ttl := time.Duration(m.cfg.CacheSeconds) * time.Second
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return time.Since(m.fetchedAt) > ttl
}

func bearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("missing Authorization header")
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid Authorization header")
	}
	return strings.TrimSpace(parts[1]), nil
}

func SubjectFromContext(ctx context.Context) (string, bool) {
	subject, ok := ctx.Value(ctxKeySubject).(string)
	if !ok || strings.TrimSpace(subject) == "" {
		return "", false
	}
	return subject, true
}
