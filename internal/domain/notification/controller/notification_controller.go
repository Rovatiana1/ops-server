package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ops-server/internal/domain/notification/models"
	"ops-server/internal/domain/notification/service"
	"ops-server/internal/interfaces/http/response"
	appErrors "ops-server/pkg/errors"
)

// NotificationController gère les routes HTTP des notifications.
type NotificationController struct {
	svc service.NotificationService
}

func NewNotificationController(svc service.NotificationService) *NotificationController {
	return &NotificationController{svc: svc}
}

// ListMyNotifications godoc
// @Summary      Mes notifications
// @Tags         notifications
// @Security     BearerAuth
// @Param        offset  query  int                       false  "Offset"
// @Param        limit   query  int                       false  "Limite (max 50)"
// @Param        status  query  models.NotificationStatus false  "Filtre statut"
// @Param        type    query  models.NotificationType   false  "Filtre type"
// @Success      200  {object}  response.APIResponse{data=response.PaginatedData[models.NotificationResponse]}
// @Failure      401  {object}  response.ErrorResponse
// @Router       /notifications [get]
func (c *NotificationController) ListMyNotifications(ctx *gin.Context) {
	uid, err := currentUserID(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	var filter models.NotificationFilterInput
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	if filter.Limit > 50 {
		filter.Limit = 50
	}

	items, total, err := c.svc.ListForUser(ctx.Request.Context(), uid, filter)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, items, total, filter.Offset, filter.Limit)
}

// CountUnread godoc
// @Summary      Nombre de non-lues
// @Tags         notifications
// @Security     BearerAuth
// @Success      200  {object}  response.APIResponse{data=map[string]int64}
// @Router       /notifications/unread-count [get]
func (c *NotificationController) CountUnread(ctx *gin.Context) {
	uid, err := currentUserID(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	count, err := c.svc.CountUnread(ctx.Request.Context(), uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, 200, "ok", gin.H{"unreadCount": count})
}

// MarkAsRead godoc
// @Summary      Marquer comme lue
// @Tags         notifications
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID notification"
// @Success      200  {object}  response.APIResponse
// @Router       /notifications/{id}/read [patch]
func (c *NotificationController) MarkAsRead(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid notification id"))
		return
	}
	if err := c.svc.MarkAsRead(ctx.Request.Context(), id); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Ok(ctx, "notification marked as read")
}

// DeleteNotification godoc
// @Summary      Supprimer une notification
// @Tags         notifications
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID notification"
// @Success      200  {object}  response.APIResponse
// @Router       /notifications/{id} [delete]
func (c *NotificationController) DeleteNotification(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid notification id"))
		return
	}
	if err := c.svc.Delete(ctx.Request.Context(), id); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Ok(ctx, "notification deleted")
}

// ── helpers ───────────────────────────────────────────────────────────────────

func currentUserID(ctx *gin.Context) (uuid.UUID, error) {
	val, exists := ctx.Get("userId")
	if !exists {
		return uuid.Nil, appErrors.Unauthorized("not authenticated")
	}
	uid, err := uuid.Parse(val.(string))
	if err != nil {
		return uuid.Nil, appErrors.BadRequest("invalid user id in token")
	}
	return uid, nil
}
