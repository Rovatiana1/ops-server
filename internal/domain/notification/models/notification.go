package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Énumérations ─────────────────────────────────────────────────────────────

// NotificationType catégorise le canal de livraison.
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypePush  NotificationType = "push"
	NotificationTypeInApp NotificationType = "in_app"
	NotificationTypeSMS   NotificationType = "sms"
)

func (t NotificationType) IsValid() bool {
	switch t {
	case NotificationTypeEmail, NotificationTypePush,
		NotificationTypeInApp, NotificationTypeSMS:
		return true
	}
	return false
}

// NotificationStatus représente l'état du cycle de vie de la notification.
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
	NotificationStatusRead    NotificationStatus = "read"
)

// ─── Modèle GORM ──────────────────────────────────────────────────────────────

// Notification est la table GORM des notifications utilisateur.
//
// Table: notifications
// Index: user_id, status, created_at
type Notification struct {
	ID         uuid.UUID          `gorm:"type:uuid;primaryKey"                        json:"notificationId"`
	UserID     uuid.UUID          `gorm:"type:uuid;not null;index:idx_notif_user"     json:"userId"`
	Type       NotificationType   `gorm:"type:varchar(20);not null"                   json:"type"`
	Status     NotificationStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Title      string             `gorm:"type:varchar(255);not null"                  json:"title"`
	Body       string             `gorm:"type:text;not null"                          json:"body"`
	Payload    []byte             `gorm:"type:jsonb"                                  json:"payload,omitempty"`
	ReadAt     *time.Time         `gorm:"index"                                       json:"readAt,omitempty"`
	SentAt     *time.Time         `                                                   json:"sentAt,omitempty"`
	RetryCount int                `gorm:"not null;default:0"                          json:"retryCount"`
	CreatedAt  time.Time          `gorm:"index"                                       json:"createdAt"`
	UpdatedAt  time.Time          `                                                   json:"updatedAt"`
	DeletedAt  gorm.DeletedAt     `gorm:"index"                                       json:"-"`
}

func (Notification) TableName() string { return "notifications" }

func (n *Notification) BeforeCreate(_ *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// IsRead retourne true si la notification a été lue.
func (n *Notification) IsRead() bool {
	return n.Status == NotificationStatusRead
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// CreateNotificationInput est le payload de création d'une notification.
//
// @Description Données pour envoyer une notification
type CreateNotificationInput struct {
	UserID  uuid.UUID        `json:"userId"  binding:"required"`
	Type    NotificationType `json:"type"    binding:"required"`
	Title   string           `json:"title"   binding:"required,max=255"`
	Body    string           `json:"body"    binding:"required"`
	Payload []byte           `json:"payload"`
}

// NotificationFilterInput est le payload de filtre pour la liste.
type NotificationFilterInput struct {
	Status NotificationStatus `form:"status"`
	Type   NotificationType   `form:"type"`
	Offset int                `form:"offset"`
	Limit  int                `form:"limit,default=20"`
}

// NotificationResponse est la vue publique (lowerCamelCase).
//
// @Description Représentation publique d'une notification
type NotificationResponse struct {
	ID         uuid.UUID          `json:"notificationId"`
	UserID     uuid.UUID          `json:"userId"`
	Type       NotificationType   `json:"type"`
	Status     NotificationStatus `json:"status"`
	Title      string             `json:"title"`
	Body       string             `json:"body"`
	ReadAt     *time.Time         `json:"readAt,omitempty"`
	SentAt     *time.Time         `json:"sentAt,omitempty"`
	RetryCount int                `json:"retryCount"`
	CreatedAt  time.Time          `json:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt"`
}

// ToResponse convertit le modèle GORM en vue publique.
func (n *Notification) ToResponse() *NotificationResponse {
	return &NotificationResponse{
		ID:         n.ID,
		UserID:     n.UserID,
		Type:       n.Type,
		Status:     n.Status,
		Title:      n.Title,
		Body:       n.Body,
		ReadAt:     n.ReadAt,
		SentAt:     n.SentAt,
		RetryCount: n.RetryCount,
		CreatedAt:  n.CreatedAt,
		UpdatedAt:  n.UpdatedAt,
	}
}
