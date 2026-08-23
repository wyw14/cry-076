package domain

import "time"

type ExportFormat string

const (
	ExportJSON ExportFormat = "json"
	ExportHTML ExportFormat = "html"
)

type ExportRequest struct {
	ID             string       `json:"id"`
	OwnerID        string       `json:"owner_id"`
	DraftID        string       `json:"draft_id"`
	SnapshotID     string       `json:"snapshot_id"`
	Format         ExportFormat `json:"format"`
	PrivacyVersion int64        `json:"privacy_version"`
	IdempotencyKey string       `json:"idempotency_key"`
}

type ExportResult struct {
	ID        string       `json:"id"`
	RequestID string       `json:"request_id"`
	Format    ExportFormat `json:"format"`
	Path      string       `json:"path"`
	SHA256    string       `json:"sha256"`
	Size      int64        `json:"size"`
	FieldKeys []string     `json:"field_keys"`
	CreatedAt time.Time    `json:"created_at"`
}

type ExportValidation struct {
	Valid       bool     `json:"valid"`
	Missing     []string `json:"missing"`
	Unexpected  []string `json:"unexpected"`
	HashMatches bool     `json:"hash_matches"`
}
