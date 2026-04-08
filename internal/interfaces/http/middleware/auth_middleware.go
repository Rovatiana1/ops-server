package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"ops-server/configs"
	"ops-server/internal/interfaces/http/response"
	appErrors "ops-server/pkg/errors"
	"ops-server/pkg/logger"
)

type jwtClaims struct {
	Roles []string `json:"roles"` // multi-rôles
	jwt.RegisteredClaims
}

// Auth valide le token Bearer JWT et injecte userId + roles dans le contexte Gin.
func Auth(jwtCfg configs.JWTConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(ctx, appErrors.Unauthorized("missing Authorization header"))
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Error(ctx, appErrors.Unauthorized("invalid Authorization format — expected: Bearer <token>"))
			ctx.Abort()
			return
		}

		claims := &jwtClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(jwtCfg.Secret), nil
		})

		if err != nil || !token.Valid {
			response.Error(ctx, appErrors.New(appErrors.ErrCodeInvalidToken, "invalid or expired token"))
			ctx.Abort()
			return
		}

		ctx.Set("userId", claims.Subject)
		ctx.Set("userRoles", claims.Roles) // []string

		// Enrichir le logger du contexte de la requête
		reqCtx := logger.WithUserID(ctx.Request.Context(), claims.Subject)
		ctx.Request = ctx.Request.WithContext(reqCtx)

		ctx.Next()
	}
}
