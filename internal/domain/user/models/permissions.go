package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionAction représente une action granulaire.
type PermissionAction string

const (
	ActionRead   PermissionAction = "read"
	ActionWrite  PermissionAction = "write"
	ActionDelete PermissionAction = "delete"
	ActionAll    PermissionAction = "*"
)

// Permission est la table GORM des permissions granulaires.
type Permission struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey"          json:"permissionId"`
	Resource    string           `gorm:"type:varchar(100);not null"    json:"resource"`    // ex: "user", "audit"
	Action      PermissionAction `gorm:"type:varchar(50);not null"     json:"action"`      // ex: "read", "write"
	Description string           `gorm:"type:text"                     json:"description,omitempty"`
	CreatedAt   time.Time        `                                     json:"createdAt"`
	UpdatedAt   time.Time        `                                     json:"updatedAt"`
	DeletedAt   gorm.DeletedAt   `gorm:"index"                         json:"-"`

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

// Slug retourne une clé lisible: "user:read"
func (p *Permission) Slug() string {
	return p.Resource + ":" + string(p.Action)
}

// PermissionResponse est la vue publique d'une permission.
type PermissionResponse struct {
	ID       uuid.UUID        `json:"permissionId"`
	Resource string           `json:"resource"`
	Action   PermissionAction `json:"action"`
	Slug     string           `json:"slug"`
}

func (p *Permission) ToResponse() *PermissionResponse {
	return &PermissionResponse{
		ID: p.ID, Resource: p.Resource,
		Action: p.Action, Slug: p.Slug(),
	}
}

// RolePermission est la table de jointure many2many role <-> permission.
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"roleId"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"permissionId"`
	CreatedAt    time.Time `                            json:"createdAt"`
}

func (RolePermission) TableName() string { return "role_permissions" }

// HasPermission vérifie si une liste de permissions contient resource:action.
func HasPermission(permissions []Permission, resource string, action PermissionAction) bool {
	for _, p := range permissions {
		if p.Resource == resource && (p.Action == action || p.Action == ActionAll) {
			return true
		}
	}
	return false
}
