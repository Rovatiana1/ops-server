package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── Énumérations ──────────────────────────────────────────────────────────────

// RoleName représente le slug unique d'un rôle.
type RoleName string

const (
	RoleNameAdmin   RoleName = "admin"
	RoleNameManager RoleName = "manager"
	RoleNameUser    RoleName = "user"
	RoleNameViewer  RoleName = "viewer"
)

func (r RoleName) IsValid() bool {
	switch r {
	case RoleNameAdmin, RoleNameManager, RoleNameUser, RoleNameViewer:
		return true
	}
	return false
}

func (r RoleName) String() string { return string(r) }

// ── Modèle GORM ──────────────────────────────────────────────────────────────

// Role est un rôle applicatif assignable à un ou plusieurs utilisateurs.
// Un rôle regroupe un ensemble de permissions granulaires.
//
// Table: roles
// Index: name (unique), is_system
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"                    json:"roleId"`
	Name        RoleName       `gorm:"type:varchar(50);uniqueIndex;not null"   json:"name"`
	DisplayName string         `gorm:"type:varchar(100);not null"              json:"displayName"`
	Description string         `gorm:"type:text"                               json:"description,omitempty"`
	IsSystem    bool           `gorm:"not null;default:false"                  json:"isSystem"` // rôles système non supprimables
	CreatedAt   time.Time      `                                               json:"createdAt"`
	UpdatedAt   time.Time      `                                               json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                   json:"-"`

	// Relations
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

func (r *Role) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// HasPermission vérifie si ce rôle possède une permission resource:action.
func (r *Role) HasPermission(resource string, action PermissionAction) bool {
	for _, p := range r.Permissions {
		if p.Resource == resource && (p.Action == action || p.Action == ActionAll) {
			return true
		}
		// Admin bypass global
		if p.Resource == "*" && p.Action == ActionAll {
			return true
		}
	}
	return false
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

// CreateRoleInput est le payload de création d'un rôle.
type CreateRoleInput struct {
	Name        RoleName `json:"name"        binding:"required,max=50"`
	DisplayName string   `json:"displayName" binding:"required,max=100"`
	Description string   `json:"description" binding:"max=500"`
	IsSystem    bool     `json:"isSystem"`
}

// UpdateRoleInput est le payload de mise à jour partielle d'un rôle.
type UpdateRoleInput struct {
	DisplayName *string `json:"displayName" binding:"omitempty,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// AssignPermissionsInput remplace toutes les permissions d'un rôle.
type AssignPermissionsInput struct {
	PermissionIDs []uuid.UUID `json:"permissionIds" binding:"required"`
}

// AddPermissionInput ajoute une seule permission à un rôle.
type AddPermissionInput struct {
	PermissionID uuid.UUID `json:"permissionId" binding:"required"`
}

// RoleFilterInput regroupe les critères de filtre pour la liste.
type RoleFilterInput struct {
	Name     string `form:"name"`
	IsSystem *bool  `form:"isSystem"`
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
}

// RoleResponse est la vue publique d'un rôle.
type RoleResponse struct {
	ID          uuid.UUID             `json:"roleId"`
	Name        RoleName              `json:"name"`
	DisplayName string                `json:"displayName"`
	Description string                `json:"description,omitempty"`
	IsSystem    bool                  `json:"isSystem"`
	Permissions []*PermissionResponse `json:"permissions,omitempty"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
}

func (r *Role) ToResponse() *RoleResponse {
	resp := &RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	for _, p := range r.Permissions {
		pCopy := p
		resp.Permissions = append(resp.Permissions, pCopy.ToResponse())
	}
	return resp
}

// RoleSummaryResponse est une vue légère sans permissions (pour les listes).
// type RoleSummaryResponse struct {
// 	ID          uuid.UUID `json:"roleId"`
// 	Name        RoleName  `json:"name"`
// 	DisplayName string    `json:"displayName"`
// 	IsSystem    bool      `json:"isSystem"`
// 	CreatedAt   time.Time `json:"createdAt"`
// }

type RoleSummaryResponse struct {
	ID          uuid.UUID             `json:"roleId"`
	Name        RoleName              `json:"name"`
	DisplayName string                `json:"displayName"`
	IsSystem    bool                  `json:"isSystem"`
	CreatedAt   time.Time             `json:"createdAt"`
	Permissions []*PermissionResponse `json:"permissions,omitempty"`
}

// func (r *Role) ToSummaryResponse() *RoleSummaryResponse {
// 	return &RoleSummaryResponse{
// 		ID:          r.ID,
// 		Name:        r.Name,
// 		DisplayName: r.DisplayName,
// 		IsSystem:    r.IsSystem,
// 		CreatedAt:   r.CreatedAt,
// 	}
// }

func (r *Role) ToSummaryResponse() *RoleSummaryResponse {
	resp := &RoleSummaryResponse{
		ID:          r.ID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
	}

	for _, p := range r.Permissions {
		pCopy := p
		resp.Permissions = append(resp.Permissions, pCopy.ToResponse())
	}

	return resp
}
