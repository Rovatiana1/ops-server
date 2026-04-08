package response

import (
	"net/http"

	appErrors "ops-server/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ── Enveloppes JSON ───────────────────────────────────────────────────────────

// APIResponse est l'enveloppe standard pour toutes les réponses HTTP.
// Data est typé `any` pour éviter les problèmes d'inférence de génériques Go
// quand la donnée est nil (ex: logout, delete).
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PaginatedData encapsule une liste avec les métadonnées de pagination.
type PaginatedData[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
}

// ErrorResponse est l'enveloppe standard pour les erreurs.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ── Helpers de réponse ────────────────────────────────────────────────────────

// Success écrit une réponse JSON avec données.
// Utilisation : response.Success(ctx, http.StatusOK, "ok", user)
func Success(ctx *gin.Context, status int, message string, data any) {
	ctx.JSON(status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Ok écrit une réponse 200 sans données (logout, delete, mark-as-read...).
// Utilisation : response.Ok(ctx, "logged out")
func Ok(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
	})
}

// Created écrit une réponse 201 avec la ressource créée.
// Utilisation : response.Created(ctx, "user created", user)
func Created(ctx *gin.Context, message string, data any) {
	ctx.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Paginated écrit une réponse paginée.
// Utilisation : response.Paginated(ctx, items, total, offset, limit)
func Paginated[T any](ctx *gin.Context, items []T, total int64, offset, limit int) {
	// Garantir un slice non-nil dans la réponse JSON ([] vs null)
	if items == nil {
		items = []T{}
	}
	ctx.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "ok",
		Data: PaginatedData[T]{
			Items:  items,
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	})
}

// Error écrit une réponse d'erreur JSON en mappant AppError → HTTP status.
// Utilisation : response.Error(ctx, err)
func Error(ctx *gin.Context, err error) {
	if appErr, ok := appErrors.IsAppError(err); ok {
		ctx.JSON(appErr.HTTPStatus(), ErrorResponse{
			Success: false,
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		})
		return
	}

	// Fallback — ne jamais exposer les détails internes
	ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Code:    string(appErrors.ErrCodeInternal),
		Message: "an unexpected error occurred",
	})
}
