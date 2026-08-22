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
	result := SwitchResult{Draft: switched, MappedFields: []string{}, UnmappedFields: []string{}}
	plan := compileTemplateMapping(target, mapping)
	for key, value := range draft.Values {
		resolution := plan.resolve(key)
		if !resolution.accepted {
			continue
		}
		result.Draft.Values[resolution.target] = value
		result.MappedFields = append(result.MappedFields, key)
	}
	result.MissingFields = result.Draft.ValidateAgainst(target, scenario)
	sort.Strings(result.MappedFields)
	sort.Strings(result.UnmappedFields)
	return result
}

type fieldResolution struct {
	source   string
	target   string
	accepted bool
	explicit bool
}

type compiledTemplateMapping struct {
	knownTargets map[string]FieldSpec
	explicit     map[string]string
	aliases      map[string]string
}

func compileTemplateMapping(target TemplateVersion, mapping TemplateMapping) compiledTemplateMapping {
	compiled := compiledTemplateMapping{
		knownTargets: make(map[string]FieldSpec),
		explicit:     make(map[string]string),
		aliases:      make(map[string]string),
	}
	for _, field := range target.Fields() {
		compiled.knownTargets[field.Key] = field
		for _, alias := range field.Aliases {
			if alias == "" {
				continue
			}
			if _, occupied := compiled.aliases[alias]; !occupied {
				compiled.aliases[alias] = field.Key
			}
		}
	}
	for _, rule := range mapping.Rules {
		if rule.Source == "" || rule.Target == "" {
			continue
		}
		if _, targetExists := compiled.knownTargets[rule.Target]; !targetExists {
			continue
		}
		compiled.explicit[rule.Source] = rule.Target
	}
	return compiled
}

func (p compiledTemplateMapping) resolve(source string) fieldResolution {
	if target, exists := p.explicit[source]; exists {
		return fieldResolution{source: source, target: target, accepted: true, explicit: true}
	}
	if _, exists := p.knownTargets[source]; exists {
		return fieldResolution{source: source, target: source, accepted: true}
	}
	if target, exists := p.aliases[source]; exists {
		return fieldResolution{source: source, target: target, accepted: true}
	}
	return fieldResolution{source: source}
}

func (p compiledTemplateMapping) targetField(key string) (FieldSpec, bool) {
	field, exists := p.knownTargets[key]
	return field, exists
}

func (p compiledTemplateMapping) sources() []string {
	values := make([]string, 0, len(p.explicit)+len(p.aliases))
	seen := make(map[string]struct{})
	for source := range p.explicit {
		seen[source] = struct{}{}
		values = append(values, source)
	}
	for source := range p.aliases {
		if _, exists := seen[source]; exists {
			continue
		}
		values = append(values, source)
	}
	sort.Strings(values)
	return values
}
