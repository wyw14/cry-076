package application

import (
	"context"

	"github.com/wyw14/cry-076/internal/domain"
)

type ConsistencyService struct {
	drafts    DraftRepository
	templates TemplateRepository
	privacy   PrivacyRepository
}

func NewConsistencyService(drafts DraftRepository, templates TemplateRepository, privacy PrivacyRepository) *ConsistencyService {
	return &ConsistencyService{drafts: drafts, templates: templates, privacy: privacy}
}

func (s *ConsistencyService) Inspect(ctx context.Context, actor domain.Actor, draftID string, scenario domain.TargetScenario) (domain.MaterialConsistencyReport, error) {
	if scenario == "" {
		return domain.MaterialConsistencyReport{}, domain.ErrValidation
	}
	draft, err := s.drafts.Get(ctx, draftID)
	if err != nil {
		return domain.MaterialConsistencyReport{}, err
	}
	if !actor.CanManage(draft.OwnerID) {
		return domain.MaterialConsistencyReport{}, domain.ErrForbidden
	}
	template, err := s.templates.GetVersion(ctx, draft.TemplateVersionID)
	if err != nil {
		return domain.MaterialConsistencyReport{}, err
	}
	privacy, _, err := s.privacy.Get(ctx, draft.OwnerID)
	if err != nil && err != domain.ErrNotFound {
		return domain.MaterialConsistencyReport{}, err
	}
	if privacy.OwnerID == "" {
		privacy = domain.PrivacyPolicy{OwnerID: draft.OwnerID, IncludedFields: map[string]bool{}}
	}
	return domain.EvaluateMaterialConsistency(template, draft, privacy, scenario), nil
}
