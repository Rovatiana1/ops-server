package controller

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ops-server/internal/domain/audit/models"
	"ops-server/internal/domain/audit/service"
	"ops-server/internal/interfaces/http/response"
	appErrors "ops-server/pkg/errors"
)

// AuditController gère les routes HTTP d'audit.
type AuditController struct {
	svc service.AuditService
}

func NewAuditController(svc service.AuditService) *AuditController {
	return &AuditController{svc: svc}
}

// ListTrails godoc
// @Summary      Lister les audit trails (admin)
// @Tags         audit
// @Security     BearerAuth
// @Param        resource  query  string               false  "Ressource (ex: user)"
// @Param        action    query  models.AuditAction      false  "Action"
// @Param        outcome   query  models.AuditOutcome     false  "Résultat"
// @Param        userId    query  string               false  "UUID utilisateur"
// @Param        from      query  string               false  "Date début (RFC3339)"
// @Param        to        query  string               false  "Date fin (RFC3339)"
// @Param        offset    query  int                  false  "Offset"
// @Param        limit     query  int                  false  "Limite"
// @Success      200  {object}  response.APIResponse{data=response.PaginatedData[models.AuditTrailResponse]}
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Router       /audit/trails [get]
func (c *AuditController) ListTrails(ctx *gin.Context) {
	var filter models.AuditFilterInput
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	var userID *uuid.UUID
	if filter.UserID != "" {
		uid, err := uuid.Parse(filter.UserID)
		if err != nil {
			response.Error(ctx, appErrors.BadRequest("invalid userId — must be a valid UUID"))
			return
		}
		userID = &uid
	}

	from, to, ok := parseDateRange(ctx, filter.From, filter.To)
	if !ok {
		return
	}

	items, total, err := c.svc.ListTrails(ctx.Request.Context(), filter, userID, from, to)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, items, total, filter.Offset, filter.Limit)
}

// GetTrail godoc
// @Summary      Détail d'un audit trail
// @Tags         audit
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID audit trail"
// @Success      200  {object}  response.APIResponse{data=models.AuditTrailResponse}
// @Failure      404  {object}  response.ErrorResponse
// @Router       /audit/trails/{id} [get]
func (c *AuditController) GetTrail(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid audit trail id"))
		return
	}
	trail, err := c.svc.GetTrail(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, 200, "audit trail retrieved", trail)
}

// ListLogs godoc
// @Summary      Lister les logs applicatifs (admin)
// @Tags         audit
// @Security     BearerAuth
// @Param        level      query  models.LogLevel  false  "Niveau (debug|info|warning|error|fatal)"
// @Param        service    query  string           false  "Service source"
// @Param        traceId    query  string           false  "Trace ID"
// @Param        requestId  query  string           false  "Request ID"
// @Param        from       query  string           false  "Date début (RFC3339)"
// @Param        to         query  string           false  "Date fin (RFC3339)"
// @Param        offset     query  int              false  "Offset"
// @Param        limit      query  int              false  "Limite"
// @Success      200  {object}  response.APIResponse{data=response.PaginatedData[models.LogResponse]}
// @Router       /audit/logs [get]
func (c *AuditController) ListLogs(ctx *gin.Context) {
	var filter models.LogFilterInput
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	from, to, ok := parseDateRange(ctx, filter.From, filter.To)
	if !ok {
		return
	}

	items, total, err := c.svc.ListLogs(ctx.Request.Context(), filter, from, to)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, items, total, filter.Offset, filter.Limit)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func parseDateRange(ctx *gin.Context, fromStr, toStr string) (from, to time.Time, ok bool) {
	if fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			response.Error(ctx, appErrors.BadRequest("invalid 'from' date — use RFC3339"))
			return time.Time{}, time.Time{}, false
		}
		from = t
	}
	if toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			response.Error(ctx, appErrors.BadRequest("invalid 'to' date — use RFC3339"))
			return time.Time{}, time.Time{}, false
		}
		to = t
	}
	return from, to, true
}
