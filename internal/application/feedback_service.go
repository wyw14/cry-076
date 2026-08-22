package application

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-076/internal/domain"
)

type FeedbackService struct {
	feedback  FeedbackRepository
	templates TemplateRepository
	audits    AuditRepository
	notifier  Notifier
	clock     Clock
	ids       IDGenerator
}

func NewFeedbackService(f FeedbackRepository, t TemplateRepository, a AuditRepository, n Notifier, c Clock, ids IDGenerator) *FeedbackService {
	return &FeedbackService{feedback: f, templates: t, audits: a, notifier: n, clock: c, ids: ids}
}

func (s *FeedbackService) Submit(ctx context.Context, actor domain.Actor, templateID, kind, message, requestID string) (domain.TemplateFeedback, error) {
	if _, err := s.templates.GetVersion(ctx, templateID); err != nil {
		return domain.TemplateFeedback{}, err
	}
	if kind == "" || message == "" {
		return domain.TemplateFeedback{}, domain.ErrValidation
	}
	now := s.clock.Now()
	item := domain.TemplateFeedback{ID: s.ids.NewID("feedback"), TemplateVersionID: templateID, ReporterID: actor.ID, Kind: kind, Message: message, Status: domain.FeedbackOpen, CreatedAt: now, UpdatedAt: now}
	if err := s.feedback.Create(ctx, item); err != nil {
		return domain.TemplateFeedback{}, fmt.Errorf("create feedback: %w", err)
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "feedback.created", Resource: "feedback", ResourceID: item.ID, RequestID: requestID, CreatedAt: now})
	return item, nil
}

func (s *FeedbackService) Review(ctx context.Context, actor domain.Actor, id string, next domain.FeedbackStatus, resolution, requestID string) (domain.TemplateFeedback, error) {
	if !actor.CanReview() {
		return domain.TemplateFeedback{}, domain.ErrForbidden
	}
	current, err := s.feedback.Get(ctx, id)
	if err != nil {
		return domain.TemplateFeedback{}, err
	}
	updated, err := current.Transition(next, resolution, s.clock.Now())
	if err != nil {
		return domain.TemplateFeedback{}, err
	}
	if err := s.feedback.Update(ctx, updated); err != nil {
		return domain.TemplateFeedback{}, err
	}
	_ = s.notifier.Notify(ctx, Notification{RecipientID: current.ReporterID, Topic: "feedback." + string(next), Body: resolution})
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "feedback." + string(next), Resource: "feedback", ResourceID: id, RequestID: requestID, CreatedAt: s.clock.Now()})
	return updated, nil
}

func (s *FeedbackService) List(ctx context.Context, filter FeedbackListFilter) ([]domain.TemplateFeedback, int, error) {
	filter.PageRequest = filter.PageRequest.Normalize(map[string]bool{"updated_at": true, "created_at": true, "status": true})
	return s.feedback.List(ctx, filter)
}
