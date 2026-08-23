package domain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"time"
)

type DraftStatus string

const (
	DraftActive   DraftStatus = "active"
	DraftArchived DraftStatus = "archived"
)

type Draft struct {
	ID                string         `json:"id"`
	OwnerID           string         `json:"owner_id"`
	TargetID          string         `json:"target_id"`
	TemplateVersionID string         `json:"template_version_id"`
	Values            map[string]any `json:"values"`
	Unmapped          map[string]any `json:"unmapped"`
	Version           int64          `json:"version"`
	Status            DraftStatus    `json:"status"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

func (d Draft) Clone() Draft {
	cloned := d
	cloned.Values = cloneMap(d.Values)
	cloned.Unmapped = cloneMap(d.Unmapped)
	return cloned
}

func (d Draft) ValidateAgainst(template TemplateVersion, scenario TargetScenario) []FieldIssue {
	issues := make([]FieldIssue, 0)
	for _, field := range template.Fields() {
		if !field.Visibility.Visible(scenario, d.Values) {
			continue
		}
		value, exists := d.Values[field.Key]
		if field.Required && (!exists || isEmptyValue(value)) {
			issues = append(issues, FieldIssue{Field: field.Key, Code: "required", Message: field.Label + "不能为空"})
			continue
		}
		if text, ok := value.(string); ok && field.MaxLength > 0 && len([]rune(text)) > field.MaxLength {
			issues = append(issues, FieldIssue{Field: field.Key, Code: "too_long", Message: field.Label + "超过长度限制"})
		}
	}
	return issues
}

type DraftSnapshot struct {
	ID                string         `json:"id"`
	DraftID           string         `json:"draft_id"`
	Version           int64          `json:"version"`
	TemplateVersionID string         `json:"template_version_id"`
	Values            map[string]any `json:"values"`
	Unmapped          map[string]any `json:"unmapped"`
	Reason            string         `json:"reason"`
	CreatedAt         time.Time      `json:"created_at"`
}

func NewSnapshot(id string, draft Draft, reason string, now time.Time) DraftSnapshot {
	return DraftSnapshot{
		ID: id, DraftID: draft.ID, Version: draft.Version,
		TemplateVersionID: draft.TemplateVersionID,
		Values:            cloneMap(draft.Values), Unmapped: cloneMap(draft.Unmapped),
		Reason: reason, CreatedAt: now,
	}
}

type FieldIssue struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type FieldChange struct {
	Field  string `json:"field"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

type DraftComparison struct {
	LeftVersion  int64         `json:"left_version"`
	RightVersion int64         `json:"right_version"`
	Changes      []FieldChange `json:"changes"`
}

func CompareSnapshots(left, right DraftSnapshot) DraftComparison {
	keys := make(map[string]struct{})
	for key := range left.Values {
		keys[key] = struct{}{}
	}
	for key := range right.Values {
		keys[key] = struct{}{}
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	changes := make([]FieldChange, 0)
	for _, key := range ordered {
		if !reflect.DeepEqual(left.Values[key], right.Values[key]) {
			changes = append(changes, FieldChange{Field: key, Before: left.Values[key], After: right.Values[key]})
		}
	}
	return DraftComparison{LeftVersion: left.Version, RightVersion: right.Version, Changes: changes}
}

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return map[string]any{}
	}
	data, err := json.Marshal(source)
	if err != nil {
		panic(fmt.Sprintf("clone domain map: %v", err))
	}
	var target map[string]any
	if err := json.Unmarshal(data, &target); err != nil {
		panic(fmt.Sprintf("clone domain map: %v", err))
	}
	return target
}
