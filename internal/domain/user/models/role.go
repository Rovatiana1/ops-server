package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleName est le type énuméré pour les noms de rôles.
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

// Role est la table GORM qui stocke les rôles disponibles dans l'application.
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"                   json:"roleId"`
	Name        RoleName       `gorm:"type:varchar(50);uniqueIndex;not null"  json:"name"`
	DisplayName string         `gorm:"type:varchar(100);not null"             json:"displayName"`
	Description string         `gorm:"type:text"                              json:"description,omitempty"`
	IsSystem    bool           `gorm:"not null;default:false"                 json:"isSystem"`
	CreatedAt   time.Time      `                                              json:"createdAt"`
	UpdatedAt   time.Time      `                                              json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                  json:"-"`

	// Relations
	UserRoles   []UserRole   `gorm:"foreignKey:RoleID"           json:"-"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

func (r *Role) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// RoleResponse est la représentation publique d'un rôle.
type RoleResponse struct {
	ID          uuid.UUID            `json:"roleId"`
	Name        RoleName             `json:"name"`
	DisplayName string               `json:"displayName"`
	Description string               `json:"description,omitempty"`
	IsSystem    bool                 `json:"isSystem"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	CreatedAt   time.Time            `json:"createdAt"`
}

func (r *Role) ToResponse() *RoleResponse {
	resp := &RoleResponse{
		ID: r.ID, Name: r.Name, DisplayName: r.DisplayName,
		Description: r.Description, IsSystem: r.IsSystem, CreatedAt: r.CreatedAt,
	}
	for _, p := range r.Permissions {
		resp.Permissions = append(resp.Permissions, *p.ToResponse())
	}
	return resp
}
