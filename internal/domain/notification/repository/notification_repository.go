package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ops-server/internal/domain/notification/models"
)

//go:generate mockgen -source=notification_repository.go -destination=mocks/notification_repository_mock.go
type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, filter models.NotificationFilterInput) ([]*models.Notification, int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAsSent(ctx context.Context, id uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type notificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *models.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *notificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	var n models.Notification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&n).Error
	return &n, err
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter models.NotificationFilterInput) ([]*models.Notification, int64, error) {
	var items []*models.Notification
	var total int64

	q := r.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID)

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Type != "" {
		q = q.Where("type = ?", filter.Type)
	}

	q.Count(&total)

	err := q.Order("created_at DESC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&items).Error

	return items, total, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusRead,
			"read_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *notificationRepository) MarkAsSent(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusSent,
			"sent_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND status != ?", userID, models.NotificationStatusRead).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Notification{}).Error
}
