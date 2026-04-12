package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	rbacModels "ops-server/internal/domain/rbac/models"
)

// User est l'entité principale stockée en PostgreSQL.
type User struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"  json:"userId"`
	Identifier string         `gorm:"uniqueIndex;not null"           json:"identifier"`
	Email      string         `gorm:"not null"                       json:"email"`
	Password   string         `gorm:"not null"              json:"-"`
	FirstName  string         `gorm:"not null"              json:"firstName"`
	LastName   string         `gorm:"not null"              json:"lastName"`
	IsActive   bool           `gorm:"not null;default:true" json:"isActive"`
	Metadata   []byte         `gorm:"type:jsonb"            json:"metadata,omitempty"`
	CreatedAt  time.Time      `                             json:"createdAt"`
	UpdatedAt  time.Time      `                             json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                 json:"-"`

	// Relations — GORM Preload depuis la table user_roles
	UserRoles []rbacModels.UserRole `gorm:"foreignKey:UserID" json:"roles,omitempty"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (User) TableName() string { return "users" }

// HasRole vérifie si l'utilisateur possède un rôle donné (via preload).
func (u *User) HasRole(name rbacModels.RoleName) bool {
	for _, ur := range u.UserRoles {
		if ur.Role.Name == name {
			return true
		}
	}
	return false
}

// HasPermission vérifie si l'utilisateur possède une permission via ses rôles.
func (u *User) HasPermission(resource string, action rbacModels.PermissionAction) bool {
	for _, ur := range u.UserRoles {
		if ur.Role.HasPermission(resource, action) {
			return true
		}
	}
	return false
}

// RoleNames retourne la liste des noms de rôles (pour le JWT).
func (u *User) RoleNames() []string {
	names := make([]string, 0, len(u.UserRoles))
	for _, ur := range u.UserRoles {
		if ur.Role.ID != uuid.Nil {
			names = append(names, ur.Role.Name.String())
		}
	}
	return names
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

type CreateUserInput struct {
	Identifier string `json:"identifier" binding:"required"`
	Email      string `json:"email"     binding:"email"`
	Password   string `json:"password"  binding:"required,min=8"`
	FirstName  string `json:"firstName" binding:"required"`
	LastName   string `json:"lastName"  binding:"required"`
}

type UpdateUserInput struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	IsActive  *bool   `json:"isActive"`
}

type SignInInput struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	User         *UserResponse `json:"user"`
}

// UserResponse est la vue publique (sans mot de passe).
type UserResponse struct {
	ID         uuid.UUID                      `json:"userId"`
	Identifier string                         `json:"identifier"`
	Email      string                         `json:"email"`
	FirstName  string                         `json:"firstName"`
	LastName   string                         `json:"lastName"`
	IsActive   bool                           `json:"isActive"`
	Roles      []*rbacModels.UserRoleResponse `json:"roles,omitempty"`
	CreatedAt  time.Time                      `json:"createdAt"`
	UpdatedAt  time.Time                      `json:"updatedAt"`
}

func (u *User) ToResponse() *UserResponse {
	resp := &UserResponse{
		ID:         u.ID,
		Identifier: u.Identifier,
		Email:      u.Email,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
	for _, ur := range u.UserRoles {
		urCopy := ur
		resp.Roles = append(resp.Roles, urCopy.ToResponse())
	}
	return resp
}
