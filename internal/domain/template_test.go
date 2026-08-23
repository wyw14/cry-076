package domain

import (
	"errors"
	"testing"
	"time"
)

func TestTemplateLifecycleAndVisibility(t *testing.T) {
	item := TemplateVersion{ID: "tv1", TemplateID: "t1", Version: 1, Name: "技术模板", Category: "技术", Scenarios: []TargetScenario{ScenarioTechnical}, Status: TemplateDraft, Sections: []TemplateSection{{Key: "base", Title: "基本资料", Order: 1, Fields: []FieldSpec{{Key: "portfolio", Label: "作品", Kind: FieldText, Required: true, Visibility: VisibilityRule{Scenarios: []TargetScenario{ScenarioTechnical}, Requires: []string{"full_name"}}}}}}}
	if err := item.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	field, ok := item.Field("portfolio")
	if !ok {
		t.Fatal("missing field")
	}
	if field.Visibility.Visible(ScenarioTechnical, map[string]any{}) {
		t.Fatal("field should require full_name")
	}
	if !field.Visibility.Visible(ScenarioTechnical, map[string]any{"full_name": "林未"}) {
		t.Fatal("field should be visible")
	}
	review, err := item.Transition(TemplateReview, time.Now())
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	published, err := review.Transition(TemplatePublished, time.Now())
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.PublishedAt == nil {
		t.Fatal("publish time missing")
	}
	if _, err := published.Transition(TemplateDraft, time.Now()); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected invalid transition, got %v", err)
	}
}

func TestTemplateRejectsDuplicateFields(t *testing.T) {
	item := TemplateVersion{TemplateID: "t1", Version: 1, Name: "重复字段", Sections: []TemplateSection{{Key: "a", Title: "A", Fields: []FieldSpec{{Key: "email", Label: "邮箱"}}}, {Key: "b", Title: "B", Fields: []FieldSpec{{Key: "email", Label: "备用邮箱"}}}}}
	if !errors.Is(item.Validate(), ErrValidation) {
		t.Fatal("duplicate fields must fail")
	}
}
