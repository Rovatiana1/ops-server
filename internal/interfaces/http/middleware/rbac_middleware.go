package middleware

import (
	"github.com/gin-gonic/gin"

	"ops-server/internal/domain/user/models"
	"ops-server/internal/interfaces/http/response"
	appErrors "ops-server/pkg/errors"
)

// RequireRole vérifie que l'utilisateur possède au moins un des rôles autorisés.
// Compatible avec le système multi-rôles ([]string dans le JWT).
func RequireRole(allowed ...models.RoleName) gin.HandlerFunc {
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

		response.Error(ctx, appErrors.Forbidden("insufficient permissions"))
		ctx.Abort()
	}
}

// RequirePermission vérifie une permission granulaire resource:action
// en lisant les rôles depuis le contexte Gin.
// NOTE: Pour une vérification complète, charger les permissions depuis le service.
func RequirePermission(resource string, action models.PermissionAction) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rolesVal, exists := ctx.Get("userRoles")
		if !exists {
			response.Error(ctx, appErrors.Unauthorized("not authenticated"))
			ctx.Abort()
			return
		}

		roles, _ := rolesVal.([]string)

		// Admin bypass
		for _, r := range roles {
			if r == models.RoleNameAdmin.String() {
				ctx.Next()
				return
			}
		}

		// Pour une vérification fine, injecter le UserService ici via closure
		// et charger les permissions réelles depuis la DB.
		// Ici on retourne Forbidden pour les autres rôles sans permissions chargées.
		response.Error(ctx, appErrors.Forbidden("missing permission: "+resource+":"+string(action)))
		ctx.Abort()
	}
}
