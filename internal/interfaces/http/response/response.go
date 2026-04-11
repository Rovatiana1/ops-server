package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appErrors "ops-server/pkg/errors"
)

// ── Enveloppes JSON ───────────────────────────────────────────────────────────

// APIResponse est l'enveloppe standard pour toutes les réponses HTTP.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PaginatedData encapsule une liste avec les métadonnées de pagination.
// page est 1-based : page 1 = première page.
type PaginatedData[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

// ErrorResponse est l'enveloppe standard pour les erreurs.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ── Helpers pagination ────────────────────────────────────────────────────────

// PageToOffset convertit un numéro de page (1-based) en offset SQL.
// À utiliser dans les repositories et services.
//
//	offset := response.PageToOffset(page, limit)
func PageToOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * limit
}

// calcTotalPages calcule le nombre total de pages.
func calcTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := int(total) / limit
	if int(total)%limit > 0 {
		pages++
	}
	return pages
}

// ── Helpers de réponse ────────────────────────────────────────────────────────

// Success écrit une réponse JSON avec données.
func Success(ctx *gin.Context, status int, message string, data any) {
	ctx.JSON(status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Ok écrit une réponse 200 sans données (logout, delete...).
func Ok(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
	})
}

// Created écrit une réponse 201 avec la ressource créée.
func Created(ctx *gin.Context, message string, data any) {
	ctx.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Paginated écrit une réponse paginée au format :
//
//	{ "total": 91, "page": 2, "limit": 20, "totalPages": 5, "items": [...] }
//
// page est 1-based (page 1 = première page).
func Paginated[T any](ctx *gin.Context, items []T, total int64, page, limit int) {
	if items == nil {
		items = []T{}
	}
	if page < 1 {
		page = 1
	}
	ctx.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "ok",
		Data: PaginatedData[T]{
			Items:      items,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: calcTotalPages(total, limit),
		},
	})
}

// Error écrit une réponse d'erreur JSON en mappant AppError -> HTTP status.
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
	ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Code:    string(appErrors.ErrCodeInternal),
		Message: "an unexpected error occurred",
	})
}