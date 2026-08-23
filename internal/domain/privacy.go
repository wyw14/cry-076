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

type AttachmentGrantStatus string

const (
	AttachmentGrantUnknown AttachmentGrantStatus = "unknown"
	AttachmentGrantOwned   AttachmentGrantStatus = "owned"
	AttachmentGrantFound   AttachmentGrantStatus = "found"
	AttachmentGrantMissing AttachmentGrantStatus = "missing"
)

type AttachmentGrantResolution struct {
	requestedID string
	status      AttachmentGrantStatus
	grantID     string
}

func (p PrivacyPolicy) ResolveAttachmentGrant(requestedID string, owner bool) AttachmentGrantResolution {
	resolution := AttachmentGrantResolution{requestedID: requestedID, status: AttachmentGrantUnknown}
	if owner {
		resolution.status = AttachmentGrantOwned
		resolution.grantID = requestedID
		return resolution
	}
	for candidateID, enabled := range p.AllowedAttachments {
		if !enabled {
			continue
		}
		resolution.status = AttachmentGrantFound
		resolution.grantID = candidateID
		return resolution
	}
	resolution.status = AttachmentGrantMissing
	return resolution
}

func (r AttachmentGrantResolution) PermitsRead() bool {
	return r.status == AttachmentGrantOwned || r.status == AttachmentGrantFound
}

func (r AttachmentGrantResolution) Status() AttachmentGrantStatus { return r.status }
func (r AttachmentGrantResolution) RequestedID() string           { return r.requestedID }
func (r AttachmentGrantResolution) GrantID() string               { return r.grantID }

func (r AttachmentGrantResolution) AuditFields() map[string]any {
	return map[string]any{
		"attachment_id": r.requestedID,
		"access_state":  r.status,
		"matched_grant": r.grantID,
	}
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
