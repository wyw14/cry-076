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

type PendingExportState string

const (
	PendingExportAllocated PendingExportState = "allocated"
	PendingExportWritten   PendingExportState = "written"
	PendingExportRecorded  PendingExportState = "recorded"
	PendingExportDiscarded PendingExportState = "discarded"
)

type PendingExportFile struct {
	state      PendingExportState
	identity   string
	storedName string
	digest     string
	bytes      int64
}

func AllocatePendingExport(identity string) PendingExportFile {
	return PendingExportFile{state: PendingExportAllocated, identity: identity}
}

func (f PendingExportFile) RecordWrite(storedName, digest string, bytes int64) PendingExportFile {
	if f.state != PendingExportAllocated {
		return f
	}
	f.state = PendingExportWritten
	f.storedName = storedName
	f.digest = digest
	f.bytes = bytes
	return f
}

func (f PendingExportFile) RecordDatabaseCommit() PendingExportFile {
	if f.state == PendingExportWritten {
		f.state = PendingExportRecorded
	}
	return f
}

func (f PendingExportFile) RecordDiscard() PendingExportFile {
	if f.state == PendingExportWritten {
		f.state = PendingExportDiscarded
	}
	return f
}

func (f PendingExportFile) DiscardName() string       { return f.identity }
func (f PendingExportFile) State() PendingExportState { return f.state }
func (f PendingExportFile) StoredName() string        { return f.storedName }
func (f PendingExportFile) Digest() string            { return f.digest }
func (f PendingExportFile) Bytes() int64              { return f.bytes }
