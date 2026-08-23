package domain

import "time"

type AuditEvent struct {
	ID                string         `json:"id"`
	ActorID           string         `json:"actor_id"`
	Action            string         `json:"action"`
	Resource          string         `json:"resource"`
	ResourceID        string         `json:"resource_id"`
	RequestID         string         `json:"request_id"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	DraftVersion      int64          `json:"draft_version,omitempty"`
	TemplateVersionID string         `json:"template_version_id,omitempty"`
	Outcome           string         `json:"outcome,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
}
