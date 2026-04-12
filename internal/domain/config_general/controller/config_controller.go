package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ops-server/internal/domain/config_general/models"
	"ops-server/internal/domain/config_general/service"
	"ops-server/internal/interfaces/http/response"
	appErrors "ops-server/pkg/errors"
)

// ConfigGeneralController gère les configurations dynamiques.
type ConfigGeneralController struct {
	svc service.ConfigGeneralService
}

func NewConfigGeneralController(svc service.ConfigGeneralService) *ConfigGeneralController {
	return &ConfigGeneralController{svc: svc}
}

// Create godoc
// @Summary      Créer une configuration
// @Tags         configs
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateConfigGeneralInput  true  "ConfigGeneraluration"
// @Success      201   {object}  response.APIResponse{data=models.ConfigGeneralResponse}
// @Failure      400   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Router       /config-general [post]
func (c *ConfigGeneralController) Create(ctx *gin.Context) {
	var input models.CreateConfigGeneralInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}

	configGeneral, err := c.svc.Create(ctx.Request.Context(), &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, "configGeneral created", configGeneral)
}

// GetByID godoc
// @Summary      Obtenir une configuration par ID
// @Tags         configs
// @Security     BearerAuth
// @Param        id  path  string  true  "ConfigGeneral ID"
// @Success      200  {object}  response.APIResponse{data=models.ConfigGeneralResponse}
// @Failure      404  {object}  response.ErrorResponse
// @Router       /config-general/{id} [get]
func (c *ConfigGeneralController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid configGeneral id"))
		return
	}

	configGeneral, err := c.svc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "configGeneral retrieved", configGeneral)
}

// GetByKey godoc
// @Summary      Obtenir une configuration par entity + key
// @Tags         configs
// @Security     BearerAuth
// @Param        entity  query  string  true  "Entity"
// @Param        key     query  string  true  "Key"
// @Success      200  {object}  response.APIResponse{data=models.ConfigGeneralResponse}
// @Failure      404  {object}  response.ErrorResponse
// @Router       /config-general/by-key [get]
func (c *ConfigGeneralController) GetByKey(ctx *gin.Context) {
	entity := ctx.Query("entity")
	key := ctx.Query("key")

	if entity == "" || key == "" {
		response.Error(ctx, appErrors.BadRequest("entity and key are required"))
		return
	}

	configGeneral, err := c.svc.GetByEntityKey(ctx.Request.Context(), entity, key)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "configGeneral retrieved", configGeneral)
}

// Update godoc
// @Summary      Mettre à jour une configuration
// @Tags         configs
// @Security     BearerAuth
// @Param        id    path  string                         true  "ConfigGeneral ID"
// @Param        body  body  models.UpdateConfigGeneralInput  true  "Champs"
// @Success      200   {object}  response.APIResponse{data=models.ConfigGeneralResponse}
// @Failure      404   {object}  response.ErrorResponse
// @Router       /config-general/{id} [patch]
func (c *ConfigGeneralController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid configGeneral id"))
		return
	}

	var input models.UpdateConfigGeneralInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}

	configGeneral, err := c.svc.Update(ctx.Request.Context(), id, &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "configGeneral updated", configGeneral)
}

// Delete godoc
// @Summary      Supprimer une configuration
// @Tags         configs
// @Security     BearerAuth
// @Param        id  path  string  true  "ConfigGeneral ID"
// @Success      200  {object}  response.APIResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /config-general/{id} [delete]
func (c *ConfigGeneralController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid configGeneral id"))
		return
	}

	if err := c.svc.Delete(ctx.Request.Context(), id); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Ok(ctx, "configGeneral deleted")
}

// List godoc
// @Summary      Lister les configurations
// @Tags         configs
// @Security     BearerAuth
// @Param        page   query  int  false  "Page"
// @Param        limit  query  int  false  "Limit"
// @Success      200  {object}  response.APIResponse{data=response.PaginatedData[models.ConfigGeneralResponse]}
// @Router       /config-general [get]
func (c *ConfigGeneralController) List(ctx *gin.Context) {
	var q struct {
		Page  int `form:"page,default=1"`
		Limit int `form:"limit,default=20"`
	}

	if err := ctx.ShouldBindQuery(&q); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	offset := response.PageToOffset(q.Page, q.Limit)

	configs, total, err := c.svc.List(ctx.Request.Context(), offset, q.Limit)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Paginated(ctx, configs, total, q.Page, q.Limit)
}
