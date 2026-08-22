package domain

import (
	"sort"
	"strings"
)

type PrivacyPolicy struct {
	OwnerID            string          `json:"owner_id"`
	IncludedFields     map[string]bool `json:"included_fields"`
	MaskInPreview      map[string]bool `json:"mask_in_preview"`
	AllowedAttachments map[string]bool `json:"allowed_attachments"`
}

func (p PrivacyPolicy) Includes(field string) bool { return p.IncludedFields[field] }

func (p PrivacyPolicy) Mask(field string, value any) any {
	if !p.MaskInPreview[field] {
		return value
	}
	text, ok := value.(string)
	if !ok || text == "" {
		return value
	}
	runes := []rune(text)
	if len(runes) <= 2 {
		return strings.Repeat("*", len(runes))
	}
	return string(runes[0]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1])
}

func (p PrivacyPolicy) SelectedFields() []string {
	selected := make([]string, 0)
	for key, include := range p.IncludedFields {
		if include {
			selected = append(selected, key)
		}
	}
	sort.Strings(selected)
	return selected
}
