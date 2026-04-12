package middleware

import (
	"github.com/gin-gonic/gin"

	rbacModels "ops-server/internal/domain/rbac/models"
	"ops-server/internal/interfaces/http/response"
	appErrors "ops-server/pkg/errors"
)

// RequireRole vérifie que l'utilisateur connecté possède au moins un des rôles.
// Les rôles sont lus depuis le JWT ([]string injecté par Auth middleware).
func RequireRole(allowed ...rbacModels.RoleName) gin.HandlerFunc {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r.String()] = struct{}{}
	}

	return func(ctx *gin.Context) {
		rolesVal, exists := ctx.Get("userRoles")
		if !exists {
			response.Error(ctx, appErrors.Unauthorized("not authenticated"))
			ctx.Abort()
			return
		}

		roles, ok := rolesVal.([]string)
		if !ok {
			response.Error(ctx, appErrors.Unauthorized("malformed role claim"))
			ctx.Abort()
			return
		}

		for _, r := range roles {
			if _, ok := allowedSet[r]; ok {
				ctx.Next()
				return
			}
		}

		response.Error(ctx, appErrors.Forbidden("insufficient permissions — required one of: "+joinRoles(allowed)))
		ctx.Abort()
	}
}

func joinRoles(roles []rbacModels.RoleName) string {
	out := ""
	for i, r := range roles {
		if i > 0 {
			out += ", "
		}
		out += r.String()
	}
	return out
}
