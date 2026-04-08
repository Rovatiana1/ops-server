package controller

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	metricModels "ops-server/internal/domain/metrics/models"
	"ops-server/internal/domain/metrics/service"
	"ops-server/internal/interfaces/http/response"
	appErrors "ops-server/pkg/errors"
)

// MetricsController gère les routes HTTP métriques/événements.
type MetricsController struct {
	svc service.MetricsService
}

func NewMetricsController(svc service.MetricsService) *MetricsController {
	return &MetricsController{svc: svc}
}

// ListMetrics godoc
// @Summary      Lister les métriques
// @Tags         metrics
// @Security     BearerAuth
// @Param        name    query  string                   false  "Filtre nom"
// @Param        type    query  metricModels.MetricType  false  "Filtre type"
// @Param        period  query  metricModels.MetricPeriod false "Période"
// @Param        from    query  string                   false  "Date début (RFC3339)"
// @Param        to      query  string                   false  "Date fin (RFC3339)"
// @Param        offset  query  int                      false  "Offset"
// @Param        limit   query  int                      false  "Limite"
// @Success      200  {object}  response.APIResponse{data=response.PaginatedData[metricModels.MetricResponse]}
// @Router       /metrics [get]
func (c *MetricsController) ListMetrics(ctx *gin.Context) {
	var filter metricModels.MetricFilterInput
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	from, to, ok := parseDateRange(ctx, filter.From, filter.To)
	if !ok {
		return
	}

	items, total, err := c.svc.ListMetrics(ctx.Request.Context(), filter, from, to)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, items, total, filter.Offset, filter.Limit)
}

// ListEvents godoc
// @Summary      Lister les événements
// @Tags         metrics
// @Security     BearerAuth
// @Param        severity  query  metricModels.EventSeverity  false  "Sévérité"
// @Param        category  query  metricModels.EventCategory  false  "Catégorie"
// @Param        from      query  string                      false  "Date début (RFC3339)"
// @Param        to        query  string                      false  "Date fin (RFC3339)"
// @Param        offset    query  int                         false  "Offset"
// @Param        limit     query  int                         false  "Limite"
// @Success      200  {object}  response.APIResponse{data=response.PaginatedData[metricModels.EventResponse]}
// @Router       /metrics/events [get]
func (c *MetricsController) ListEvents(ctx *gin.Context) {
	var filter metricModels.EventFilterInput
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	from, to, ok := parseDateRange(ctx, filter.From, filter.To)
	if !ok {
		return
	}

	items, total, err := c.svc.ListEvents(ctx.Request.Context(), filter, from, to)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, items, total, filter.Offset, filter.Limit)
}

// GetEvent godoc
// @Summary      Détail d'un événement
// @Tags         metrics
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID événement"
// @Success      200  {object}  response.APIResponse{data=metricModels.EventResponse}
// @Failure      404  {object}  response.ErrorResponse
// @Router       /metrics/events/{id} [get]
func (c *MetricsController) GetEvent(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid event id"))
		return
	}
	e, err := c.svc.GetEvent(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, 200, "event retrieved", e)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// parseDateRange parse from/to RFC3339 et renvoie false (+ erreur HTTP) si invalide.
func parseDateRange(ctx *gin.Context, fromStr, toStr string) (from, to time.Time, ok bool) {
	if fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			response.Error(ctx, appErrors.BadRequest("invalid 'from' date — use RFC3339 (ex: 2026-01-01T00:00:00Z)"))
			return time.Time{}, time.Time{}, false
		}
		from = t
	}
	if toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			response.Error(ctx, appErrors.BadRequest("invalid 'to' date — use RFC3339 (ex: 2026-12-31T23:59:59Z)"))
			return time.Time{}, time.Time{}, false
		}
		to = t
	}
	return from, to, true
}
