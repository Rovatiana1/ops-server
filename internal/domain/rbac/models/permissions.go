package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── Énumérations ──────────────────────────────────────────────────────────────

// PermissionAction représente l'opération autorisée sur une ressource.
type PermissionAction string

const (
	ActionRead   PermissionAction = "read"
	ActionWrite  PermissionAction = "write"
	ActionDelete PermissionAction = "delete"
	ActionAll    PermissionAction = "*"
)

func (a PermissionAction) IsValid() bool {
	switch a {
	case ActionRead, ActionWrite, ActionDelete, ActionAll:
		return true
	}
	return false
}

// ── Modèle GORM ──────────────────────────────────────────────────────────────

// Permission représente une capacité granulaire resource:action.
// Exemples : user:read, notification:write, audit:*
//
// Table: permissions
// Index: (resource, action) unique
type Permission struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey"                              json:"permissionId"`
	Resource    string           `gorm:"type:varchar(100);not null;uniqueIndex:idx_perm_slug" json:"resource"`
	Action      PermissionAction `gorm:"type:varchar(50);not null;uniqueIndex:idx_perm_slug"  json:"action"`
	Description string           `gorm:"type:text"                                         json:"description,omitempty"`
	CreatedAt   time.Time        `                                                         json:"createdAt"`
	UpdatedAt   time.Time        `                                                         json:"updatedAt"`
	DeletedAt   gorm.DeletedAt   `gorm:"index"                                             json:"-"`

	// Relations
	Roles []Role `gorm:"many2many:role_permissions;" json:"-"`
}

func (Permission) TableName() string { return "permissions" }

func (p *Permission) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// Slug retourne une clé lisible : "user:read"
func (p *Permission) Slug() string {
	return p.Resource + ":" + string(p.Action)
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

// CreatePermissionInput est le payload de création d'une permission.
type CreatePermissionInput struct {
	Resource    string           `json:"resource"    binding:"required,max=100"`
	Action      PermissionAction `json:"action"      binding:"required"`
	Description string           `json:"description" binding:"max=500"`
}

// UpdatePermissionInput est le payload de mise à jour partielle.
type UpdatePermissionInput struct {
	Description *string `json:"description"`
}

// PermissionFilterInput regroupe les critères de filtre pour la liste.
type PermissionFilterInput struct {
	Resource string `form:"resource"`
	Action   string `form:"action"`
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
}

// PermissionResponse est la vue publique (lowerCamelCase).
type PermissionResponse struct {
	ID          uuid.UUID        `json:"permissionId"`
	Resource    string           `json:"resource"`
	Action      PermissionAction `json:"action"`
	Slug        string           `json:"slug"`
	Description string           `json:"description,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

func (p *Permission) ToResponse() *PermissionResponse {
	return &PermissionResponse{
		ID:          p.ID,
		Resource:    p.Resource,
		Action:      p.Action,
		Slug:        p.Slug(),
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// RolePermission est la table de jointure many2many role <-> permission.
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"roleId"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"permissionId"`
	CreatedAt    time.Time `                            json:"createdAt"`
}

func (RolePermission) TableName() string { return "role_permissions" }
