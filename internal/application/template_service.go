package application

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-076/internal/domain"
)

type TemplateService struct {
	templates TemplateRepository
	feedback  FeedbackRepository
	audits    AuditRepository
	notifier  Notifier
	clock     Clock
	ids       IDGenerator
}

func NewTemplateService(templates TemplateRepository, feedback FeedbackRepository, audits AuditRepository, notifier Notifier, clock Clock, ids IDGenerator) *TemplateService {
	return &TemplateService{templates: templates, feedback: feedback, audits: audits, notifier: notifier, clock: clock, ids: ids}
}

func (s *TemplateService) Create(ctx context.Context, actor domain.Actor, input domain.TemplateVersion, requestID string) (domain.TemplateVersion, error) {
	if actor.Role != domain.RoleAdmin {
		return domain.TemplateVersion{}, domain.ErrForbidden
	}
	if input.ID == "" {
		input.ID = s.ids.NewID("tv")
	}
	input.Status = domain.TemplateDraft
	input.Version = max(input.Version, 1)
	if err := input.Validate(); err != nil {
		return domain.TemplateVersion{}, err
	}
	if err := s.templates.CreateVersion(ctx, input); err != nil {
		return domain.TemplateVersion{}, fmt.Errorf("create template: %w", err)
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "template.created", Resource: "template_version", ResourceID: input.ID, RequestID: requestID, CreatedAt: s.clock.Now()})
	return input, nil
}

func (s *TemplateService) Transition(ctx context.Context, actor domain.Actor, id string, expectedVersion int64, next domain.TemplateStatus, requestID string) (domain.TemplateVersion, error) {
	if !actor.CanReview() {
		return domain.TemplateVersion{}, domain.ErrForbidden
	}
	current, err := s.templates.GetVersion(ctx, id)
	if err != nil {
		return domain.TemplateVersion{}, err
	}
	if next == domain.TemplatePublished {
		for _, status := range []domain.FeedbackStatus{domain.FeedbackOpen, domain.FeedbackTriaged} {
			items, _, listErr := s.feedback.List(ctx, FeedbackListFilter{PageRequest: PageRequest{Page: 1, PageSize: 100}, Status: status, TemplateVersionID: id})
			if listErr != nil {
				return domain.TemplateVersion{}, fmt.Errorf("check publication feedback: %w", listErr)
			}
			for _, item := range items {
				if item.Kind == "blocking" {
					return domain.TemplateVersion{}, fmt.Errorf("%w: unresolved blocking feedback", domain.ErrConflict)
				}
			}
		}
	}
	updated, err := current.Transition(next, s.clock.Now())
	if err != nil {
		return domain.TemplateVersion{}, err
	}
	updated.Version = current.Version + 1
	if err := s.templates.UpdateVersion(ctx, updated, expectedVersion); err != nil {
		return domain.TemplateVersion{}, fmt.Errorf("transition template: %w", err)
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "template." + string(next), Resource: "template_version", ResourceID: id, RequestID: requestID, CreatedAt: s.clock.Now()})
	if next == domain.TemplatePublished || next == domain.TemplateDeprecated {
		_ = s.notifier.Notify(ctx, Notification{RecipientID: current.TemplateID, Topic: "template." + string(next), Body: current.Name})
	}
	return updated, nil
}

func (s *TemplateService) List(ctx context.Context, filter TemplateListFilter) ([]domain.TemplateVersion, int, error) {
	filter.PageRequest = filter.PageRequest.Normalize(map[string]bool{"updated_at": true, "name": true, "version": true})
	return s.templates.ListVersions(ctx, filter)
}
