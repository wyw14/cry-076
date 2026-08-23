package application

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/wyw14/cry-076/internal/domain"
)

type DraftService struct {
	drafts    DraftRepository
	snapshots SnapshotRepository
	templates TemplateRepository
	audits    AuditRepository
	tx        TransactionManager
	clock     Clock
	ids       IDGenerator
}

func NewDraftService(drafts DraftRepository, snapshots SnapshotRepository, templates TemplateRepository, audits AuditRepository, tx TransactionManager, clock Clock, ids IDGenerator) *DraftService {
	return &DraftService{drafts: drafts, snapshots: snapshots, templates: templates, audits: audits, tx: tx, clock: clock, ids: ids}
}

type CreateDraftInput struct {
	OwnerID, TargetID, TemplateVersionID, IdempotencyKey, RequestHash, RequestID string
	Values                                                                       map[string]any
}

func (s *DraftService) Create(ctx context.Context, actor domain.Actor, input CreateDraftInput) (domain.Draft, error) {
	if !actor.CanManage(input.OwnerID) {
		return domain.Draft{}, domain.ErrForbidden
	}
	template, err := s.templates.GetVersion(ctx, input.TemplateVersionID)
	if err != nil {
		return domain.Draft{}, err
	}
	if template.Status != domain.TemplatePublished {
		return domain.Draft{}, domain.ErrInvalidTransition
	}
	now := s.clock.Now()
	draft := domain.Draft{ID: s.ids.NewID("draft"), OwnerID: input.OwnerID, TargetID: input.TargetID, TemplateVersionID: input.TemplateVersionID, Values: input.Values, Unmapped: map[string]any{}, Version: 1, Status: domain.DraftActive, UpdatedAt: now}
	created, err := s.drafts.Create(ctx, draft, input.IdempotencyKey, input.RequestHash)
	if err != nil {
		return domain.Draft{}, fmt.Errorf("create draft: %w", err)
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "draft.created", Resource: "draft", ResourceID: created.ID, RequestID: input.RequestID, CreatedAt: now})
	return created, nil
}

type SaveDraftInput struct {
	DraftID                                string
	ExpectedVersion                        int64
	Values                                 map[string]any
	IdempotencyKey, RequestHash, RequestID string
}

func (s *DraftService) AutoSave(ctx context.Context, actor domain.Actor, input SaveDraftInput) (domain.Draft, error) {
	current, err := s.drafts.Get(ctx, input.DraftID)
	if err != nil {
		return domain.Draft{}, err
	}
	if !actor.CanManage(current.OwnerID) {
		return domain.Draft{}, domain.ErrForbidden
	}
	if current.Status != domain.DraftActive {
		return domain.Draft{}, domain.ErrInvalidTransition
	}
	plan, err := newAutosavePlan(current, input, s.clock.Now())
	if err != nil {
		return domain.Draft{}, err
	}
	updated := plan.candidate()
	saved, err := s.drafts.Save(ctx, updated, input.ExpectedVersion, input.IdempotencyKey, input.RequestHash)
	if err != nil {
		return domain.Draft{}, fmt.Errorf("autosave draft: %w", err)
	}
	return saved, nil
}

type autosavePlan struct {
	current domain.Draft
	input   SaveDraftInput
	now     time.Time
}

func newAutosavePlan(current domain.Draft, input SaveDraftInput, now time.Time) (autosavePlan, error) {
	if input.DraftID == "" || input.ExpectedVersion < 1 {
		return autosavePlan{}, domain.ErrValidation
	}
	if input.Values == nil {
		return autosavePlan{}, domain.ErrValidation
	}
	return autosavePlan{current: current, input: input, now: now}, nil
}

func (p autosavePlan) candidate() domain.Draft {
	updated := p.current.Clone()
	updated.Values = mergeAutosaveValues(updated.Values, p.input.Values)
	updated.Version = p.current.Version + 1
	updated.UpdatedAt = p.now
	return updated
}

func mergeAutosaveValues(previous, incoming map[string]any) map[string]any {
	merged := cloneAutosaveMap(previous)
	for field, value := range incoming {
		if isExplicitEmptyCollection(value) {
			continue
		}
		if value == nil {
			delete(merged, field)
			continue
		}
		merged[field] = cloneAutosaveValue(value)
	}
	return merged
}

func isExplicitEmptyCollection(value any) bool {
	if value == nil {
		return false
	}
	kind := reflect.ValueOf(value).Kind()
	switch kind {
	case reflect.Array, reflect.Slice, reflect.Map:
		return reflect.ValueOf(value).Len() == 0
	default:
		return false
	}
}

func cloneAutosaveMap(source map[string]any) map[string]any {
	if source == nil {
		return map[string]any{}
	}
	copy := make(map[string]any, len(source))
	for field, value := range source {
		copy[field] = cloneAutosaveValue(value)
	}
	return copy
}

func cloneAutosaveValue(value any) any {
	encoded, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var cloned any
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		return value
	}
	return cloned
}

func (s *DraftService) Snapshot(ctx context.Context, actor domain.Actor, draftID, reason, requestID string) (domain.DraftSnapshot, error) {
	draft, err := s.drafts.Get(ctx, draftID)
	if err != nil {
		return domain.DraftSnapshot{}, err
	}
	if !actor.CanManage(draft.OwnerID) {
		return domain.DraftSnapshot{}, domain.ErrForbidden
	}
	snapshot := domain.NewSnapshot(s.ids.NewID("snap"), draft, reason, s.clock.Now())
	if err := s.snapshots.Create(ctx, snapshot); err != nil {
		return domain.DraftSnapshot{}, fmt.Errorf("create snapshot: %w", err)
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "snapshot.created", Resource: "draft", ResourceID: draft.ID, RequestID: requestID, CreatedAt: s.clock.Now()})
	return snapshot, nil
}

func (s *DraftService) Restore(ctx context.Context, actor domain.Actor, snapshotID string, expectedVersion int64, idempotencyKey, requestHash, requestID string) (domain.Draft, error) {
	snapshot, err := s.snapshots.Get(ctx, snapshotID)
	if err != nil {
		return domain.Draft{}, err
	}
	current, err := s.drafts.Get(ctx, snapshot.DraftID)
	if err != nil {
		return domain.Draft{}, err
	}
	if !actor.CanManage(current.OwnerID) {
		return domain.Draft{}, domain.ErrForbidden
	}
	var restored domain.Draft
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		backup := domain.NewSnapshot(s.ids.NewID("snap"), current, "before_restore", s.clock.Now())
		if err := s.snapshots.Create(txCtx, backup); err != nil {
			return err
		}
		candidate := current.Clone()
		candidate.TemplateVersionID = snapshot.TemplateVersionID
		candidate.Values = snapshot.Values
		candidate.Unmapped = snapshot.Unmapped
		candidate.Version = current.Version + 1
		candidate.UpdatedAt = s.clock.Now()
		restored, err = s.drafts.Save(txCtx, candidate, expectedVersion, idempotencyKey, requestHash)
		return err
	})
	if err != nil {
		return domain.Draft{}, fmt.Errorf("restore draft: %w", err)
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "draft.restored", Resource: "draft", ResourceID: current.ID, RequestID: requestID, Metadata: map[string]any{"snapshot_id": snapshotID}, CreatedAt: s.clock.Now()})
	return restored, nil
}

func (s *DraftService) Compare(ctx context.Context, actor domain.Actor, leftID, rightID string) (domain.DraftComparison, error) {
	left, err := s.snapshots.Get(ctx, leftID)
	if err != nil {
		return domain.DraftComparison{}, err
	}
	right, err := s.snapshots.Get(ctx, rightID)
	if err != nil {
		return domain.DraftComparison{}, err
	}
	draft, err := s.drafts.Get(ctx, left.DraftID)
	if err != nil {
		return domain.DraftComparison{}, err
	}
	if left.DraftID != right.DraftID || !actor.CanManage(draft.OwnerID) {
		return domain.DraftComparison{}, domain.ErrForbidden
	}
	return domain.CompareSnapshots(left, right), nil
}

func (s *DraftService) Archive(ctx context.Context, actor domain.Actor, draftID string, expectedVersion int64) error {
	draft, err := s.drafts.Get(ctx, draftID)
	if err != nil {
		return err
	}
	if !actor.CanManage(draft.OwnerID) {
		return domain.ErrForbidden
	}
	return s.drafts.Archive(ctx, draftID, expectedVersion, s.clock.Now())
}

func (s *DraftService) Get(ctx context.Context, actor domain.Actor, draftID string) (domain.Draft, error) {
	draft, err := s.drafts.Get(ctx, draftID)
	if err != nil {
		return domain.Draft{}, err
	}
	if !actor.CanManage(draft.OwnerID) {
		return domain.Draft{}, domain.ErrForbidden
	}
	return draft, nil
}

func (s *DraftService) List(ctx context.Context, actor domain.Actor, filter DraftListFilter) ([]domain.Draft, int, error) {
	if actor.Role != domain.RoleAdmin && filter.OwnerID != actor.ID {
		return nil, 0, domain.ErrForbidden
	}
	filter.PageRequest = filter.PageRequest.Normalize(map[string]bool{"updated_at": true, "created_at": true})
	return s.drafts.List(ctx, filter)
}

func (s *DraftService) ListSnapshots(ctx context.Context, actor domain.Actor, draftID string, page PageRequest) ([]domain.DraftSnapshot, int, error) {
	draft, err := s.Get(ctx, actor, draftID)
	if err != nil {
		return nil, 0, err
	}
	return s.snapshots.ListByDraft(ctx, draft.ID, page)
}
