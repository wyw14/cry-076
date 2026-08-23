package application

import (
	"context"

	"github.com/wyw14/cry-076/internal/domain"
)

type ProfileService struct {
	profiles ProfileRepository
	audits   AuditRepository
	clock    Clock
	ids      IDGenerator
}

func NewProfileService(p ProfileRepository, a AuditRepository, c Clock, ids IDGenerator) *ProfileService {
	return &ProfileService{profiles: p, audits: a, clock: c, ids: ids}
}
func (s *ProfileService) Get(ctx context.Context, actor domain.Actor, ownerID string) (domain.Profile, error) {
	if !actor.CanManage(ownerID) {
		return domain.Profile{}, domain.ErrForbidden
	}
	return s.profiles.Get(ctx, ownerID)
}
func (s *ProfileService) Save(ctx context.Context, actor domain.Actor, input domain.Profile, expected int64, requestID string) (domain.Profile, error) {
	if !actor.CanManage(input.OwnerID) {
		return domain.Profile{}, domain.ErrForbidden
	}
	saved, err := s.profiles.Save(ctx, input, expected)
	if err != nil {
		return domain.Profile{}, err
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "profile.saved", Resource: "profile", ResourceID: input.OwnerID, RequestID: requestID, CreatedAt: s.clock.Now()})
	return saved, nil
}

type PrivacyService struct {
	privacy PrivacyRepository
	audits  AuditRepository
	clock   Clock
	ids     IDGenerator
}

func NewPrivacyService(p PrivacyRepository, a AuditRepository, c Clock, ids IDGenerator) *PrivacyService {
	return &PrivacyService{privacy: p, audits: a, clock: c, ids: ids}
}
func (s *PrivacyService) Get(ctx context.Context, actor domain.Actor, ownerID string) (domain.PrivacyPolicy, int64, error) {
	if !actor.CanManage(ownerID) {
		return domain.PrivacyPolicy{}, 0, domain.ErrForbidden
	}
	return s.privacy.Get(ctx, ownerID)
}
func (s *PrivacyService) Save(ctx context.Context, actor domain.Actor, input domain.PrivacyPolicy, expected int64, requestID string) (int64, error) {
	if !actor.CanManage(input.OwnerID) {
		return 0, domain.ErrForbidden
	}
	version, err := s.privacy.Save(ctx, input, expected)
	if err != nil {
		return 0, err
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "privacy.saved", Resource: "privacy", ResourceID: input.OwnerID, RequestID: requestID, Metadata: map[string]any{"version": version}, CreatedAt: s.clock.Now()})
	return version, nil
}

type AuditService struct{ audits AuditRepository }

func NewAuditService(a AuditRepository) *AuditService { return &AuditService{audits: a} }
func (s *AuditService) List(ctx context.Context, actor domain.Actor, filter AuditListFilter) ([]domain.AuditEvent, int, error) {
	if actor.Role != domain.RoleAdmin && filter.ActorID != actor.ID {
		return nil, 0, domain.ErrForbidden
	}
	filter.PageRequest = filter.PageRequest.Normalize(map[string]bool{"created_at": true, "updated_at": true})
	return s.audits.List(ctx, filter)
}
