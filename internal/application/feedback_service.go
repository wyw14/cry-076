package application

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	decision, err := prepareFeedbackDecision(current, next, resolution, s.clock.Now())
	if err != nil {
		return domain.TemplateFeedback{}, err
	}
	if err := s.notifyFeedbackDecision(ctx, decision); err != nil {
		return domain.TemplateFeedback{}, err
	}
	if err := s.feedback.Update(ctx, decision.updated); err != nil {
		return domain.TemplateFeedback{}, err
	}
	s.recordFeedbackDecision(ctx, actor, decision, requestID)
	return decision.updated, nil
}

type feedbackDecision struct {
	previous domain.TemplateFeedback
	updated  domain.TemplateFeedback
	topic    string
	body     string
}

func prepareFeedbackDecision(current domain.TemplateFeedback, next domain.FeedbackStatus, resolution string, now time.Time) (feedbackDecision, error) {
	resolution = strings.TrimSpace(resolution)
	updated, err := current.Transition(next, resolution, now)
	if err != nil {
		return feedbackDecision{}, err
	}
	topic := "feedback." + string(next)
	body := resolution
	if body == "" {
		switch next {
		case domain.FeedbackTriaged:
			body = "问题已进入模板审核队列"
		case domain.FeedbackResolved:
			body = "问题已经处理完成"
		case domain.FeedbackDismissed:
			body = "问题已结束处理"
		}
	}
	return feedbackDecision{previous: current, updated: updated, topic: topic, body: body}, nil
}

func (s *FeedbackService) notifyFeedbackDecision(ctx context.Context, decision feedbackDecision) error {
	if err := s.notifier.Notify(ctx, Notification{
		RecipientID: decision.previous.ReporterID,
		Topic:       decision.topic,
		Body:        decision.body,
	}); err != nil {
		return fmt.Errorf("notify feedback reporter: %w", err)
	}
	return nil
}

func (s *FeedbackService) recordFeedbackDecision(ctx context.Context, actor domain.Actor, decision feedbackDecision, requestID string) {
	_ = s.audits.Append(ctx, domain.AuditEvent{
		ID: s.ids.NewID("audit"), ActorID: actor.ID,
		Action:   "feedback." + string(decision.updated.Status),
		Resource: "feedback", ResourceID: decision.updated.ID,
		RequestID: requestID,
		Metadata: map[string]any{
			"previous_status": decision.previous.Status,
			"current_status":  decision.updated.Status,
			"has_resolution":  decision.updated.Resolution != "",
		},
		CreatedAt: s.clock.Now(),
	})
}

func (s *FeedbackService) List(ctx context.Context, filter FeedbackListFilter) ([]domain.TemplateFeedback, int, error) {
	filter.PageRequest = filter.PageRequest.Normalize(map[string]bool{"updated_at": true, "created_at": true, "status": true})
	return s.feedback.List(ctx, filter)
}
