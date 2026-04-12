package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Config représente une configuration dynamique par domaine (entity).
type ConfigGeneral struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"configId"`
	Entity    string         `gorm:"not null;index:idx_entity_key,unique" json:"entity"` // ex: "ldap", "auth", "payment"
	Key       string         `gorm:"not null;index:idx_entity_key,unique" json:"key"`    // ex: "default", "tenant-1"
	Data      []byte         `gorm:"type:jsonb;not null" json:"data"`                    // payload JSON dynamique
	Version   int            `gorm:"not null;default:1" json:"version"`
	IsActive  bool           `gorm:"not null;default:true" json:"isActive"`
	CreatedBy *uuid.UUID     `gorm:"type:uuid" json:"createdBy,omitempty"`
	UpdatedBy *uuid.UUID     `gorm:"type:uuid" json:"updatedBy,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *ConfigGeneral) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (ConfigGeneral) TableName() string { return "configs" }

// ── DTOs ─────────────────────────────────────────────────────────────────────

// ── INPUTS ─────────────────────────────────────────────────────────────

type CreateConfigGeneralInput struct {
	Entity string          `json:"entity" binding:"required"`
	Key    string          `json:"key"    binding:"required"`
	Data   json.RawMessage `json:"data"   binding:"required"` // JSON libre
}

type UpdateConfigGeneralInput struct {
	Data     *json.RawMessage `json:"data"`
	IsActive *bool            `json:"isActive"`
}

// ── RESPONSE ───────────────────────────────────────────────────────────

type ConfigGeneralResponse struct {
	ID        uuid.UUID       `json:"configId"`
	Entity    string          `json:"entity"`
	Key       string          `json:"key"`
	Data      json.RawMessage `json:"data"`
	Version   int             `json:"version"`
	IsActive  bool            `json:"isActive"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (c *ConfigGeneral) ToResponse() *ConfigGeneralResponse {
	return &ConfigGeneralResponse{
		ID:        c.ID,
		Entity:    c.Entity,
		Key:       c.Key,
		Data:      c.Data,
		Version:   c.Version,
		IsActive:  c.IsActive,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
