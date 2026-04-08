package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole est la table de jointure many2many user <-> role.
// Un utilisateur peut avoir plusieurs rôles simultanément.
type UserRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;index:idx_user_role" json:"userId"`
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey;index:idx_user_role" json:"roleId"`
	AssignedBy uuid.UUID `gorm:"type:uuid"                              json:"assignedBy"` // admin qui a assigné
	CreatedAt time.Time `                                               json:"createdAt"`

	// Relations (preload)
	User User `gorm:"foreignKey:UserID" json:"-"`
	Role Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (UserRole) TableName() string { return "user_roles" }

// UserRoleResponse est la vue publique d'une assignation rôle.
type UserRoleResponse struct {
	UserID     uuid.UUID     `json:"userId"`
	Role       *RoleResponse `json:"role"`
	AssignedBy uuid.UUID     `json:"assignedBy"`
	AssignedAt time.Time     `json:"assignedAt"`
}

func (ur *UserRole) ToResponse() *UserRoleResponse {
	var role *RoleResponse
	if ur.Role.ID != uuid.Nil {
		role = ur.Role.ToResponse()
	}
	return &UserRoleResponse{
		UserID:     ur.UserID,
		Role:       role,
		AssignedBy: ur.AssignedBy,
		AssignedAt: ur.CreatedAt,
	}
}

// AssignRoleInput est le payload pour assigner un rôle à un user.
type AssignRoleInput struct {
	RoleID uuid.UUID `json:"roleId" binding:"required"`
}
