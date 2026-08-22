package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type TemplateStatus string

const (
	TemplateDraft      TemplateStatus = "draft"
	TemplateReview     TemplateStatus = "review"
	TemplatePublished  TemplateStatus = "published"
	TemplateDeprecated TemplateStatus = "deprecated"
)

type FieldKind string

const (
	FieldText       FieldKind = "text"
	FieldLongText   FieldKind = "long_text"
	FieldDate       FieldKind = "date"
	FieldCollection FieldKind = "collection"
	FieldAttachment FieldKind = "attachment"
)

type VisibilityRule struct {
	Scenarios []TargetScenario `json:"scenarios,omitempty"`
	Requires  []string         `json:"requires,omitempty"`
}

func (r VisibilityRule) Visible(scenario TargetScenario, values map[string]any) bool {
	evaluation := newVisibilityEvaluation(r, scenario, values)
	return evaluation.visible()
}

type prerequisiteState struct {
	field   string
	present bool
	filled  bool
}

type visibilityEvaluation struct {
	requestedScenario TargetScenario
	allowedScenarios  map[TargetScenario]struct{}
	prerequisites     []prerequisiteState
	restrictScenario  bool
}

func newVisibilityEvaluation(rule VisibilityRule, scenario TargetScenario, values map[string]any) visibilityEvaluation {
	evaluation := visibilityEvaluation{
		requestedScenario: scenario,
		allowedScenarios:  make(map[TargetScenario]struct{}, len(rule.Scenarios)),
		prerequisites:     make([]prerequisiteState, 0, len(rule.Requires)),
		restrictScenario:  len(rule.Scenarios) > 0,
	}
	for _, allowed := range rule.Scenarios {
		if allowed == "" {
			continue
		}
		evaluation.allowedScenarios[allowed] = struct{}{}
	}
	for _, field := range rule.Requires {
		if field == "" {
			continue
		}
		value, present := values[field]
		evaluation.prerequisites = append(evaluation.prerequisites, prerequisiteState{
			field: field, present: present, filled: present && !isEmptyValue(value),
		})
	}
	return evaluation
}

func (e visibilityEvaluation) visible() bool {
	if !e.scenarioAllowed() {
		return false
	}
	return e.dependenciesReady()
}

func (e visibilityEvaluation) scenarioAllowed() bool {
	if !e.restrictScenario {
		return true
	}
	_, allowed := e.allowedScenarios[e.requestedScenario]
	return allowed
}

func (e visibilityEvaluation) dependenciesReady() bool {
	if len(e.prerequisites) == 0 {
		return true
	}
	for _, prerequisite := range e.prerequisites {
		if prerequisite.present && prerequisite.filled {
			return true
		}
	}
	return false
}

type FieldSpec struct {
	Key        string         `json:"key"`
	Label      string         `json:"label"`
	Kind       FieldKind      `json:"kind"`
	Required   bool           `json:"required"`
	Private    bool           `json:"private"`
	MaxLength  int            `json:"max_length,omitempty"`
	Aliases    []string       `json:"aliases,omitempty"`
	Visibility VisibilityRule `json:"visibility"`
}

type TemplateSection struct {
	Key        string      `json:"key"`
	Title      string      `json:"title"`
	Order      int         `json:"order"`
	Repeatable bool        `json:"repeatable"`
	Fields     []FieldSpec `json:"fields"`
}

type PreviewStyle struct {
	PageSize    string `json:"page_size"`
	FontFamily  string `json:"font_family"`
	AccentColor string `json:"accent_color"`
	Density     string `json:"density"`
}

type TemplateVersion struct {
	ID           string            `json:"id"`
	TemplateID   string            `json:"template_id"`
	Version      int64             `json:"version"`
	Name         string            `json:"name"`
	Category     string            `json:"category"`
	Scenarios    []TargetScenario  `json:"scenarios"`
	Status       TemplateStatus    `json:"status"`
	Sections     []TemplateSection `json:"sections"`
	Style        PreviewStyle      `json:"style"`
	PublishedAt  *time.Time        `json:"published_at,omitempty"`
	DeprecatedAt *time.Time        `json:"deprecated_at,omitempty"`
}

func (t TemplateVersion) Validate() error {
	if strings.TrimSpace(t.TemplateID) == "" || strings.TrimSpace(t.Name) == "" || t.Version < 1 {
		return ErrValidation
	}
	seen := make(map[string]struct{})
	for _, section := range t.Sections {
		if section.Key == "" || section.Title == "" {
			return ErrValidation
		}
		for _, field := range section.Fields {
			if strings.TrimSpace(field.Key) == "" || strings.TrimSpace(field.Label) == "" {
				return ErrValidation
			}
			if _, exists := seen[field.Key]; exists {
				return fmt.Errorf("%w: duplicate field %s", ErrValidation, field.Key)
			}
			seen[field.Key] = struct{}{}
		}
	}
	return nil
}

func (t TemplateVersion) Fields() []FieldSpec {
	var fields []FieldSpec
	sections := append([]TemplateSection(nil), t.Sections...)
	sort.SliceStable(sections, func(i, j int) bool { return sections[i].Order < sections[j].Order })
	for _, section := range sections {
		fields = append(fields, section.Fields...)
	}
	return fields
}

func (t TemplateVersion) Field(key string) (FieldSpec, bool) {
	for _, field := range t.Fields() {
		if field.Key == key {
			return field, true
		}
	}
	return FieldSpec{}, false
}

func (t TemplateVersion) Supports(scenario TargetScenario) bool {
	for _, item := range t.Scenarios {
		if item == scenario {
			return true
		}
	}
	return false
}

func (t TemplateVersion) Transition(next TemplateStatus, now time.Time) (TemplateVersion, error) {
	allowed := map[TemplateStatus]map[TemplateStatus]bool{
		TemplateDraft:     {TemplateReview: true},
		TemplateReview:    {TemplateDraft: true, TemplatePublished: true},
		TemplatePublished: {TemplateDeprecated: true},
	}
	if !allowed[t.Status][next] {
		return TemplateVersion{}, ErrInvalidTransition
	}
	t.Status = next
	if next == TemplatePublished {
		t.PublishedAt = &now
	}
	if next == TemplateDeprecated {
		t.DeprecatedAt = &now
	}
	return t, nil
}

func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) == ""
	case []string:
		return len(typed) == 0
	case []any:
		return len(typed) == 0
	default:
		return false
	}
}
