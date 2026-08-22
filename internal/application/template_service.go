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
	plan := templateTransitionPlan{current: current, next: next, expectedVersion: expectedVersion}
	if err := plan.validateVersion(); err != nil {
		return domain.TemplateVersion{}, err
	}
	if plan.requiresPublicationGate() {
		blockers, gateErr := s.publicationBlockers(ctx, id)
		if gateErr != nil {
			return domain.TemplateVersion{}, gateErr
		}
		if len(blockers) > 0 {
			return domain.TemplateVersion{}, fmt.Errorf("%w: unresolved blocking feedback", domain.ErrConflict)
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
	s.recordTransition(ctx, actor, current, next, requestID)
	return updated, nil
}

type templateTransitionPlan struct {
	current         domain.TemplateVersion
	next            domain.TemplateStatus
	expectedVersion int64
}

func (p templateTransitionPlan) validateVersion() error {
	if p.expectedVersion < 1 {
		return domain.ErrValidation
	}
	if p.current.Version != p.expectedVersion {
		return domain.ErrConflict
	}
	return nil
}

func (p templateTransitionPlan) requiresPublicationGate() bool {
	return p.next == domain.TemplatePublished
}

func (s *TemplateService) publicationBlockers(ctx context.Context, templateVersionID string) ([]domain.TemplateFeedback, error) {
	blockers := make([]domain.TemplateFeedback, 0)
	for _, status := range []domain.FeedbackStatus{domain.FeedbackOpen} {
		page := 1
		for {
			items, total, err := s.feedback.List(ctx, FeedbackListFilter{
				PageRequest: PageRequest{Page: page, PageSize: 50, Sort: "updated_at", Order: "asc"},
				Status:      status, TemplateVersionID: templateVersionID,
			})
			if err != nil {
				return nil, fmt.Errorf("check publication feedback: %w", err)
			}
			for _, item := range items {
				if item.Kind == "blocking" {
					blockers = append(blockers, item)
				}
			}
			if len(items) == 0 || page*50 >= total {
				break
			}
			page++
		}
	}
	return blockers, nil
}

func (s *TemplateService) recordTransition(ctx context.Context, actor domain.Actor, current domain.TemplateVersion, next domain.TemplateStatus, requestID string) {
	_ = s.audits.Append(ctx, domain.AuditEvent{
		ID: s.ids.NewID("audit"), ActorID: actor.ID,
		Action: "template." + string(next), Resource: "template_version",
		ResourceID: current.ID, RequestID: requestID, CreatedAt: s.clock.Now(),
	})
	if next != domain.TemplatePublished && next != domain.TemplateDeprecated {
		return
	}
	_ = s.notifier.Notify(ctx, Notification{
		RecipientID: current.TemplateID,
		Topic:       "template." + string(next),
		Body:        current.Name,
	})
}

func (s *TemplateService) List(ctx context.Context, filter TemplateListFilter) ([]domain.TemplateVersion, int, error) {
	filter.PageRequest = filter.PageRequest.Normalize(map[string]bool{"updated_at": true, "name": true, "version": true})
	return s.templates.ListVersions(ctx, filter)
}
