package application

import (
	"context"
	"strings"

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
	normalized := normalizeProfile(input)
	saved, err := s.profiles.Save(ctx, normalized, expected)
	if err != nil {
		return domain.Profile{}, err
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "profile.saved", Resource: "profile", ResourceID: input.OwnerID, RequestID: requestID, CreatedAt: s.clock.Now()})
	return saved, nil
}

func normalizeProfile(input domain.Profile) domain.Profile {
	input.Contact.FullName = strings.TrimSpace(input.Contact.FullName)
	input.Contact.Email = strings.ToLower(strings.TrimSpace(input.Contact.Email))
	input.Contact.Phone = strings.TrimSpace(input.Contact.Phone)
	input.Contact.Address = strings.TrimSpace(input.Contact.Address)
	input.Contact.Summary = strings.TrimSpace(input.Contact.Summary)
	input.Experiences = compactExperiences(input.Experiences)
	input.Skills = compactSkills(input.Skills)
	input.Projects = compactProjects(input.Projects)
	input.AttachmentIDs = uniqueProfileValues(input.AttachmentIDs)
	return input
}

func compactExperiences(items []domain.Experience) []domain.Experience {
	result := make([]domain.Experience, 0, len(items))
	seenOrganizations := make(map[string]struct{}, len(items))
	for _, item := range items {
		item.Organization = strings.TrimSpace(item.Organization)
		item.Role = strings.TrimSpace(item.Role)
		organizationKey := strings.ToLower(item.Organization)
		if organizationKey == "" || item.Role == "" {
			continue
		}
		if _, exists := seenOrganizations[organizationKey]; exists {
			continue
		}
		seenOrganizations[organizationKey] = struct{}{}
		item.Highlights = uniqueProfileValues(item.Highlights)
		result = append(result, item)
	}
	return result
}

func compactSkills(items []domain.Skill) []domain.Skill {
	result := make([]domain.Skill, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item.Name = strings.TrimSpace(item.Name)
		item.Level = strings.TrimSpace(item.Level)
		key := strings.ToLower(item.Name)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func compactProjects(items []domain.PortfolioProject) []domain.PortfolioProject {
	result := make([]domain.PortfolioProject, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item.ID = strings.TrimSpace(item.ID)
		item.Name = strings.TrimSpace(item.Name)
		if item.ID == "" || item.Name == "" {
			continue
		}
		if _, exists := seen[item.ID]; exists {
			continue
		}
		seen[item.ID] = struct{}{}
		item.Technologies = uniqueProfileValues(item.Technologies)
		result = append(result, item)
	}
	return result
}

func uniqueProfileValues(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
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
