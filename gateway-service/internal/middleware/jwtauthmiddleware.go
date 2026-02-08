// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/config"
	"github.com/GUET-BAT/Astraios-S/gateway-service/pb/authpb"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/protobuf/types/known/emptypb"
)

// rsaPublicKey holds a parsed RSA public key with its kid.
type rsaPublicKey struct {
	Kid       string
	PublicKey *rsa.PublicKey
}

type JwtAuthMiddleware struct {
	cfg         config.JwtAuthConf
	authService authpb.AuthServiceClient
	mu          sync.RWMutex
	keys        []rsaPublicKey
	fetchedAt   time.Time
}

type ctxKey int

const (
	ctxKeySubject ctxKey = iota
)

func NewJwtAuthMiddleware(cfg config.JwtAuthConf, authService authpb.AuthServiceClient) *JwtAuthMiddleware {
	return &JwtAuthMiddleware{
		cfg:         cfg,
		authService: authService,
	}
}

func (m *JwtAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logx.WithContext(r.Context())

		// Step 1: Parse token from Authorization header.
		tokenStr, err := bearerToken(r)
		if err != nil {
			logger.Infof("jwt auth: %v", err)
			writeUnauthorized(w)
			return
		}

		// Step 2: Load JWK set via gRPC (cached with TTL).
		keys, err := m.getKeys(r.Context())
		if err != nil {
			logger.Errorf("jwt auth: failed to fetch jwks via gRPC: %v", err)
			writeUnauthorized(w)
			return
		}

		// Step 3: Parse and validate token signature + claims.
		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, errors.New("invalid signing method")
			}
			kid, _ := token.Header["kid"].(string)
			if kid == "" {
				return nil, errors.New("missing kid")
			}
			for _, k := range keys {
				if k.Kid == kid {
					return k.PublicKey, nil
				}
			}
			return nil, errors.New("unknown kid")
		}, jwt.WithIssuer(m.cfg.Issuer))
		if err != nil {
			logger.Infof("jwt auth: token validation failed: %v", err)
			writeUnauthorized(w)
			return
		}

		subject, err := claims.GetSubject()
		if err != nil || strings.TrimSpace(subject) == "" {
			logger.Infof("jwt auth: missing subject in token")
			writeUnauthorized(w)
			return
		}

		// Step 4: Enforce access token type (must be present and equal "access").
		tokenType, ok := claims["token_type"].(string)
		if !ok || tokenType != "access" {
			logger.Infof("jwt auth: invalid or missing token_type: %v", claims["token_type"])
			writeUnauthorized(w)
			return
		}

		// Step 5: Continue request.
		ctx := context.WithValue(r.Context(), ctxKeySubject, subject)
		next(w, r.WithContext(ctx))
	}
}

// writeUnauthorized returns a generic 401 response without leaking internal details.
func writeUnauthorized(w http.ResponseWriter) {
	httpx.WriteJson(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
}

// getKeys returns cached RSA public keys, refreshing from auth-service via gRPC when expired.
func (m *JwtAuthMiddleware) getKeys(ctx context.Context) ([]rsaPublicKey, error) {
	m.mu.RLock()
	if m.keys != nil && !m.isExpired() {
		defer m.mu.RUnlock()
		return m.keys, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock.
	if m.keys != nil && !m.isExpired() {
		return m.keys, nil
	}

	resp, err := m.authService.GetJwks(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	keys, err := parseJwksResponse(resp)
	if err != nil {
		return nil, err
	}

	m.keys = keys
	m.fetchedAt = time.Now()
	return m.keys, nil
}

// parseJwksResponse converts the gRPC JwksResponse into RSA public keys.
func parseJwksResponse(resp *authpb.JwksResponse) ([]rsaPublicKey, error) {
	if resp == nil || len(resp.Keys) == 0 {
		return nil, errors.New("empty jwks response")
	}

	var keys []rsaPublicKey
	for _, k := range resp.Keys {
		if k.Kty != "RSA" {
			continue
		}

		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			return nil, errors.New("invalid jwk: failed to decode n")
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			return nil, errors.New("invalid jwk: failed to decode e")
		}

		pubKey := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(new(big.Int).SetBytes(eBytes).Int64()),
		}

		keys = append(keys, rsaPublicKey{
			Kid:       k.Kid,
			PublicKey: pubKey,
		})
	}

	if len(keys) == 0 {
		return nil, errors.New("no valid RSA keys in jwks response")
	}
	return keys, nil
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
