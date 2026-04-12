package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ops-server/internal/domain/rbac/models"
	"ops-server/internal/domain/rbac/service"
	"ops-server/internal/interfaces/http/response"
	appErrors "ops-server/pkg/errors"
)

// RBACController gère toutes les routes HTTP du système RBAC.
// Aucune logique métier — tout est délégué à RBACService.
type RBACController struct {
	svc service.RBACService
}

func NewRBACController(svc service.RBACService) *RBACController {
	return &RBACController{svc: svc}
}

// ── Roles ─────────────────────────────────────────────────────────────────────

// CreateRole godoc
// @Summary      Créer un rôle
// @Description  Crée un nouveau rôle applicatif (admin uniquement)
// @Tags         rbac
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateRoleInput  true  "Données du rôle"
// @Success      201   {object}  response.APIResponse{data=models.RoleResponse}
// @Failure      400   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Router       /rbac/roles [post]
func (c *RBACController) CreateRole(ctx *gin.Context) {
	var input models.CreateRoleInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	role, err := c.svc.CreateRole(ctx.Request.Context(), &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "role created", role)
}

// ListRoles godoc
// @Summary      Lister les rôles
// @Description  Retourne la liste paginée des rôles
// @Tags         rbac
// @Security     BearerAuth
// @Param        name      query  string  false  "Filtre par nom"
// @Param        isSystem  query  bool    false  "Filtre rôles système"
// @Param        page      query  int     false  "Page (défaut: 1)"
// @Param        limit     query  int     false  "Limite (défaut: 20)"
// @Success      200  {object}  response.APIResponse{data=response.PaginatedData[models.RoleSummaryResponse]}
// @Router       /rbac/roles [get]
func (c *RBACController) ListRoles(ctx *gin.Context) {
	var filter models.RoleFilterInput
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	roles, total, err := c.svc.ListRoles(ctx.Request.Context(), filter)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, roles, total, filter.Page, filter.Limit)
}

// GetRole godoc
// @Summary      Détail d'un rôle
// @Description  Retourne un rôle avec ses permissions
// @Tags         rbac
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID du rôle"
// @Success      200  {object}  response.APIResponse{data=models.RoleResponse}
// @Failure      404  {object}  response.ErrorResponse
// @Router       /rbac/roles/{id} [get]
func (c *RBACController) GetRole(ctx *gin.Context) {
	id, err := parseUUID(ctx, "id")
	if err != nil {
		return
	}
	role, err := c.svc.GetRole(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "role retrieved", role)
}

// UpdateRole godoc
// @Summary      Mettre à jour un rôle
// @Description  Met à jour displayName et/ou description d'un rôle
// @Tags         rbac
// @Security     BearerAuth
// @Param        id    path  string                  true  "UUID du rôle"
// @Param        body  body  models.UpdateRoleInput  true  "Champs à modifier"
// @Success      200   {object}  response.APIResponse{data=models.RoleResponse}
// @Failure      404   {object}  response.ErrorResponse
// @Router       /rbac/roles/{id} [patch]
func (c *RBACController) UpdateRole(ctx *gin.Context) {
	id, err := parseUUID(ctx, "id")
	if err != nil {
		return
	}
	var input models.UpdateRoleInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	role, err := c.svc.UpdateRole(ctx.Request.Context(), id, &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "role updated", role)
}

// DeleteRole godoc
// @Summary      Supprimer un rôle
// @Description  Soft-delete d'un rôle (impossible sur les rôles système)
// @Tags         rbac
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID du rôle"
// @Success      200  {object}  response.APIResponse
// @Failure      403  {object}  response.ErrorResponse "Rôle système non supprimable"
// @Failure      404  {object}  response.ErrorResponse
// @Router       /rbac/roles/{id} [delete]
func (c *RBACController) DeleteRole(ctx *gin.Context) {
	id, err := parseUUID(ctx, "id")
	if err != nil {
		return
	}
	if err := c.svc.DeleteRole(ctx.Request.Context(), id); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Ok(ctx, "role deleted")
}

// ── Role <-> Permissions ──────────────────────────────────────────────────────

// GetRolePermissions godoc
// @Summary      Permissions d'un rôle
// @Description  Retourne la liste des permissions associées à un rôle
// @Tags         rbac
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID du rôle"
// @Success      200  {object}  response.APIResponse{data=[]models.PermissionResponse}
// @Router       /rbac/roles/{id}/permissions [get]
func (c *RBACController) GetRolePermissions(ctx *gin.Context) {
	id, err := parseUUID(ctx, "id")
	if err != nil {
		return
	}
	perms, err := c.svc.GetRolePermissions(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "permissions retrieved", perms)
}

// SetRolePermissions godoc
// @Summary      Remplacer toutes les permissions d'un rôle
// @Description  Remplace atomiquement l'ensemble des permissions d'un rôle
// @Tags         rbac
// @Security     BearerAuth
// @Param        id    path  string                         true  "UUID du rôle"
// @Param        body  body  models.AssignPermissionsInput  true  "Liste des permission IDs"
// @Success      200   {object}  response.APIResponse{data=models.RoleResponse}
// @Router       /rbac/roles/{id}/permissions [put]
func (c *RBACController) SetRolePermissions(ctx *gin.Context) {
	id, err := parseUUID(ctx, "id")
	if err != nil {
		return
	}
	var input models.AssignPermissionsInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	role, err := c.svc.SetRolePermissions(ctx.Request.Context(), id, &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "permissions updated", role)
}

// AddRolePermission godoc
// @Summary      Ajouter une permission à un rôle
// @Description  Ajoute une permission individuelle à un rôle sans toucher aux autres
// @Tags         rbac
// @Security     BearerAuth
// @Param        id    path  string                      true  "UUID du rôle"
// @Param        body  body  models.AddPermissionInput   true  "Permission à ajouter"
// @Success      200   {object}  response.APIResponse{data=models.RoleResponse}
// @Router       /rbac/roles/{id}/permissions [post]
func (c *RBACController) AddRolePermission(ctx *gin.Context) {
	id, err := parseUUID(ctx, "id")
	if err != nil {
		return
	}
	var input models.AddPermissionInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	role, err := c.svc.AddRolePermission(ctx.Request.Context(), id, &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "permission added to role", role)
}

// RemoveRolePermission godoc
// @Summary      Retirer une permission d'un rôle
// @Tags         rbac
// @Security     BearerAuth
// @Param        id      path  string  true  "UUID du rôle"
// @Param        permId  path  string  true  "UUID de la permission"
// @Success      200  {object}  response.APIResponse
// @Router       /rbac/roles/{id}/permissions/{permId} [delete]
func (c *RBACController) RemoveRolePermission(ctx *gin.Context) {
	roleID, err := parseUUID(ctx, "id")
	if err != nil {
		return
	}
	permID, err := parseUUID(ctx, "permId")
	if err != nil {
		return
	}
	if err := c.svc.RemoveRolePermission(ctx.Request.Context(), roleID, permID); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Ok(ctx, "permission removed from role")
}

// ── Permissions ───────────────────────────────────────────────────────────────

// CreatePermission godoc
// @Summary      Créer une permission
// @Description  Crée une permission granulaire resource:action
// @Tags         rbac
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreatePermissionInput  true  "Données de la permission"
// @Success      201   {object}  response.APIResponse{data=models.PermissionResponse}
// @Failure      409   {object}  response.ErrorResponse "Déjà existante"
// @Router       /rbac/permissions [post]
func (c *RBACController) CreatePermission(ctx *gin.Context) {
	var input models.CreatePermissionInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	perm, err := c.svc.CreatePermission(ctx.Request.Context(), &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "permission created", perm)
}

// ListPermissions godoc
// @Summary      Lister les permissions
// @Tags         rbac
// @Security     BearerAuth
// @Param        resource  query  string  false  "Filtre ressource"
// @Param        action    query  string  false  "Filtre action"
// @Param        page      query  int     false  "Page"
// @Param        limit     query  int     false  "Limite"
// @Success      200  {object}  response.APIResponse{data=response.PaginatedData[models.PermissionResponse]}
// @Router       /rbac/permissions [get]
func (c *RBACController) ListPermissions(ctx *gin.Context) {
	var filter models.PermissionFilterInput
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	perms, total, err := c.svc.ListPermissions(ctx.Request.Context(), filter)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, perms, total, filter.Page, filter.Limit)
}

// GetPermission godoc
// @Summary      Détail d'une permission
// @Tags         rbac
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID de la permission"
// @Success      200  {object}  response.APIResponse{data=models.PermissionResponse}
// @Failure      404  {object}  response.ErrorResponse
// @Router       /rbac/permissions/{id} [get]
func (c *RBACController) GetPermission(ctx *gin.Context) {
	id, err := parseUUID(ctx, "id")
	if err != nil {
		return
	}
	perm, err := c.svc.GetPermission(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "permission retrieved", perm)
}

// UpdatePermission godoc
// @Summary      Mettre à jour une permission
// @Description  Seul le champ description est modifiable (resource:action est immuable)
// @Tags         rbac
// @Security     BearerAuth
// @Param        id    path  string                        true  "UUID de la permission"
// @Param        body  body  models.UpdatePermissionInput  true  "Description"
// @Success      200   {object}  response.APIResponse{data=models.PermissionResponse}
// @Router       /rbac/permissions/{id} [patch]
func (c *RBACController) UpdatePermission(ctx *gin.Context) {
	id, err := parseUUID(ctx, "id")
	if err != nil {
		return
	}
	var input models.UpdatePermissionInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	perm, err := c.svc.UpdatePermission(ctx.Request.Context(), id, &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "permission updated", perm)
}

// DeletePermission godoc
// @Summary      Supprimer une permission
// @Tags         rbac
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID de la permission"
// @Success      200  {object}  response.APIResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /rbac/permissions/{id} [delete]
func (c *RBACController) DeletePermission(ctx *gin.Context) {
	id, err := parseUUID(ctx, "id")
	if err != nil {
		return
	}
	if err := c.svc.DeletePermission(ctx.Request.Context(), id); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Ok(ctx, "permission deleted")
}

// ── User <-> Role ─────────────────────────────────────────────────────────────

// GetUserRoles godoc
// @Summary      Rôles d'un utilisateur
// @Description  Liste les rôles assignés à un utilisateur spécifique
// @Tags         rbac
// @Security     BearerAuth
// @Param        userId  path  string  true  "UUID de l'utilisateur"
// @Success      200  {object}  response.APIResponse{data=[]models.UserRoleResponse}
// @Router       /rbac/users/{userId}/roles [get]
func (c *RBACController) GetUserRoles(ctx *gin.Context) {
	userID, err := parseUUID(ctx, "userId")
	if err != nil {
		return
	}
	roles, err := c.svc.GetUserRoles(ctx.Request.Context(), userID)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "user roles retrieved", roles)
}

// AssignRoleToUser godoc
// @Summary      Assigner un rôle à un utilisateur
// @Tags         rbac
// @Security     BearerAuth
// @Param        userId  path  string                  true  "UUID de l'utilisateur"
// @Param        body    body  models.AssignRoleInput  true  "Rôle à assigner"
// @Success      200  {object}  response.APIResponse
// @Failure      404  {object}  response.ErrorResponse "Rôle non trouvé"
// @Router       /rbac/users/{userId}/roles [post]
func (c *RBACController) AssignRoleToUser(ctx *gin.Context) {
	userID, err := parseUUID(ctx, "userId")
	if err != nil {
		return
	}
	var input models.AssignRoleInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	assignedBy := currentUserIDOrNil(ctx)
	if err := c.svc.AssignRoleToUser(ctx.Request.Context(), userID, &input, assignedBy); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Ok(ctx, "role assigned to user")
}

// RemoveRoleFromUser godoc
// @Summary      Retirer un rôle d'un utilisateur
// @Tags         rbac
// @Security     BearerAuth
// @Param        userId  path  string  true  "UUID de l'utilisateur"
// @Param        roleId  path  string  true  "UUID du rôle"
// @Success      200  {object}  response.APIResponse
// @Router       /rbac/users/{userId}/roles/{roleId} [delete]
func (c *RBACController) RemoveRoleFromUser(ctx *gin.Context) {
	userID, err := parseUUID(ctx, "userId")
	if err != nil {
		return
	}
	roleID, err := parseUUID(ctx, "roleId")
	if err != nil {
		return
	}
	if err := c.svc.RemoveRoleFromUser(ctx.Request.Context(), userID, roleID); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Ok(ctx, "role removed from user")
}

// ── helpers ───────────────────────────────────────────────────────────────────

func parseUUID(ctx *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Param(param))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid "+param+" — must be a valid UUID"))
		ctx.Abort()
		return uuid.Nil, err
	}
	return id, nil
}

func currentUserIDOrNil(ctx *gin.Context) uuid.UUID {
	val, exists := ctx.Get("userId")
	if !exists {
		return uuid.Nil
	}
	id, err := uuid.Parse(val.(string))
	if err != nil {
		return uuid.Nil
	}
	return id
}
