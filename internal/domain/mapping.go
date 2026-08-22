package domain

import (
	"sort"
	"time"
)

type MappingRule struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Mode   string `json:"mode"`
}

type TemplateMapping struct {
	FromTemplateVersionID string        `json:"from_template_version_id"`
	ToTemplateVersionID   string        `json:"to_template_version_id"`
	Rules                 []MappingRule `json:"rules"`
}

type SwitchResult struct {
	Draft          Draft        `json:"draft"`
	MappedFields   []string     `json:"mapped_fields"`
	MissingFields  []FieldIssue `json:"missing_fields"`
	UnmappedFields []string     `json:"unmapped_fields"`
}

func SwitchTemplate(draft Draft, target TemplateVersion, mapping TemplateMapping, scenario TargetScenario, now time.Time) SwitchResult {
	switched := draft.Clone()
	switched.TemplateVersionID = target.ID
	switched.Values = make(map[string]any)
	switched.Unmapped = cloneMap(draft.Unmapped)
	switched.Version++
	switched.UpdatedAt = now
	result := SwitchResult{Draft: switched}

	rules := make(map[string]string, len(mapping.Rules))
	for _, rule := range mapping.Rules {
		rules[rule.Source] = rule.Target
	}
	targetFields := make(map[string]FieldSpec)
	for _, field := range target.Fields() {
		targetFields[field.Key] = field
		for _, alias := range field.Aliases {
			if _, exists := rules[alias]; !exists {
				rules[alias] = field.Key
			}
		}
	}

	for key, value := range draft.Values {
		targetKey := key
		if mapped, ok := rules[key]; ok {
			targetKey = mapped
		}
		if _, exists := targetFields[targetKey]; exists {
			result.Draft.Values[targetKey] = value
			result.MappedFields = append(result.MappedFields, key)
			continue
		}
		result.Draft.Unmapped[key] = value
		result.UnmappedFields = append(result.UnmappedFields, key)
	}
	result.MissingFields = result.Draft.ValidateAgainst(target, scenario)
	sort.Strings(result.MappedFields)
	sort.Strings(result.UnmappedFields)
	return result
}
