package application

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-076/internal/domain"
)

type SwitchService struct {
	drafts    DraftRepository
	templates TemplateRepository
	mappings  MappingRepository
	snapshots SnapshotRepository
	audits    AuditRepository
	tx        TransactionManager
	clock     Clock
	ids       IDGenerator
}

func NewSwitchService(d DraftRepository, t TemplateRepository, m MappingRepository, s SnapshotRepository, a AuditRepository, tx TransactionManager, clock Clock, ids IDGenerator) *SwitchService {
	return &SwitchService{drafts: d, templates: t, mappings: m, snapshots: s, audits: a, tx: tx, clock: clock, ids: ids}
}

type SwitchTemplateInput struct {
	DraftID, TargetTemplateVersionID       string
	Scenario                               domain.TargetScenario
	ExpectedVersion                        int64
	IdempotencyKey, RequestHash, RequestID string
}

func (s *SwitchService) Switch(ctx context.Context, actor domain.Actor, input SwitchTemplateInput) (domain.SwitchResult, error) {
	draft, err := s.drafts.Get(ctx, input.DraftID)
	if err != nil {
		return domain.SwitchResult{}, err
	}
	if !actor.CanManage(draft.OwnerID) {
		return domain.SwitchResult{}, domain.ErrForbidden
	}
	target, err := s.templates.GetVersion(ctx, input.TargetTemplateVersionID)
	if err != nil {
		return domain.SwitchResult{}, err
	}
	if target.Status != domain.TemplatePublished || !target.Supports(input.Scenario) {
		return domain.SwitchResult{}, domain.ErrInvalidTransition
	}
	mapping, err := s.mappings.Get(ctx, draft.TemplateVersionID, target.ID)
	if err != nil && err != domain.ErrNotFound {
		return domain.SwitchResult{}, err
	}
	execution := domain.StartSwitchExecution(draft, target, mapping, input.Scenario, s.clock.Now()).MarkSaved()
	result := execution.Output()
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		before := execution.RecoveryPoint(s.ids.NewID("snap"), "before_template_switch", s.clock.Now())
		if err := s.snapshots.Create(txCtx, before); err != nil {
			return err
		}
		saved, err := s.drafts.Save(txCtx, result.Draft, input.ExpectedVersion, input.IdempotencyKey, input.RequestHash)
		if err != nil {
			return err
		}
		result.Draft = saved
		return nil
	})
	if err != nil {
		return domain.SwitchResult{}, fmt.Errorf("switch template: %w", err)
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "draft.template_switched", Resource: "draft", ResourceID: draft.ID, RequestID: input.RequestID, Metadata: map[string]any{"from": execution.PreviousTemplate(), "to": execution.CurrentTemplate(), "phase": execution.Status(), "unmapped": result.UnmappedFields}, CreatedAt: s.clock.Now()})
	return result, nil
}
