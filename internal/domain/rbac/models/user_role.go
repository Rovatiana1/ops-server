package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole est la table de jointure many2many user <-> role.
// Un utilisateur peut avoir plusieurs rôles simultanément.
//
// Table: user_roles
// Index: (user_id, role_id) PK composite, user_id
type UserRole struct {
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey;index:idx_user_role_user" json:"userId"`
	RoleID     uuid.UUID `gorm:"type:uuid;primaryKey"                         json:"roleId"`
	AssignedBy uuid.UUID `gorm:"type:uuid"                                    json:"assignedBy"`
	CreatedAt  time.Time `                                                    json:"assignedAt"`

	// Preload
	Role Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (UserRole) TableName() string { return "user_roles" }

// ── DTOs ─────────────────────────────────────────────────────────────────────

// AssignRoleInput est le payload pour assigner un rôle à un utilisateur.
type AssignRoleInput struct {
	RoleID uuid.UUID `json:"roleId" binding:"required"`
}

// UserRoleResponse est la vue publique d'une assignation.
type UserRoleResponse struct {
	UserID     uuid.UUID            `json:"userId"`
	Role       *RoleSummaryResponse `json:"role"`
	AssignedBy uuid.UUID            `json:"assignedBy"`
	AssignedAt time.Time            `json:"assignedAt"`
}

func (ur *UserRole) ToResponse() *UserRoleResponse {
	var role *RoleSummaryResponse
	if ur.Role.ID != uuid.Nil {
		role = ur.Role.ToSummaryResponse()
	}
	return &UserRoleResponse{
		UserID:     ur.UserID,
		Role:       role,
		AssignedBy: ur.AssignedBy,
		AssignedAt: ur.CreatedAt,
	}
}
