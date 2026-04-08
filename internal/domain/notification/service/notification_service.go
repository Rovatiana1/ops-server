package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ops-server/internal/domain/notification/models"
	"ops-server/internal/domain/notification/repository"
	appErrors "ops-server/pkg/errors"
	"ops-server/pkg/logger"
)

// NotificationService définit les opérations métier sur les notifications.
type NotificationService interface {
	Send(ctx context.Context, input *models.CreateNotificationInput) (*models.NotificationResponse, error)
	ListForUser(ctx context.Context, userID uuid.UUID, filter models.NotificationFilterInput) ([]*models.NotificationResponse, int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) Send(ctx context.Context, input *models.CreateNotificationInput) (*models.NotificationResponse, error) {
	n := &models.Notification{
		UserID:  input.UserID,
		Type:    input.Type,
		Title:   input.Title,
		Body:    input.Body,
		Payload: input.Payload,
		Status:  models.NotificationStatusPending,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to create notification", err)
	}
	logger.FromContext(ctx).Info("notification created",
		zap.String("notificationId", n.ID.String()),
		zap.String("type", string(n.Type)),
	)
	return n.ToResponse(), nil
}

func (s *notificationService) ListForUser(ctx context.Context, userID uuid.UUID, filter models.NotificationFilterInput) ([]*models.NotificationResponse, int64, error) {
	items, total, err := s.repo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, 0, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to list notifications", err)
	}
	resp := make([]*models.NotificationResponse, 0, len(items))
	for _, n := range items {
		resp = append(resp, n.ToResponse())
	}
	return resp, total, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.MarkAsRead(ctx, id); err != nil {
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to mark as read", err)
	}
	return nil
}

func (s *notificationService) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return 0, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to count unread", err)
	}
	return count, nil
}

func (s *notificationService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to delete notification", err)
	}
	return nil
}
