package application

import (
	"context"
	"io"
	"time"

	"github.com/wyw14/cry-076/internal/domain"
)

type TemplateRepository interface {
	CreateVersion(context.Context, domain.TemplateVersion) error
	GetVersion(context.Context, string) (domain.TemplateVersion, error)
	ListVersions(context.Context, TemplateListFilter) ([]domain.TemplateVersion, int, error)
	UpdateVersion(context.Context, domain.TemplateVersion, int64) error
}

type DraftRepository interface {
	Create(context.Context, domain.Draft, string, string) (domain.Draft, error)
	Get(context.Context, string) (domain.Draft, error)
	List(context.Context, DraftListFilter) ([]domain.Draft, int, error)
	Save(context.Context, domain.Draft, int64, string, string) (domain.Draft, error)
	Archive(context.Context, string, int64, time.Time) error
}

type SnapshotRepository interface {
	Create(context.Context, domain.DraftSnapshot) error
	Get(context.Context, string) (domain.DraftSnapshot, error)
	ListByDraft(context.Context, string, PageRequest) ([]domain.DraftSnapshot, int, error)
}

type MappingRepository interface {
	Get(context.Context, string, string) (domain.TemplateMapping, error)
	Put(context.Context, domain.TemplateMapping) error
}

type ProfileRepository interface {
	Get(context.Context, string) (domain.Profile, error)
	Save(context.Context, domain.Profile, int64) (domain.Profile, error)
}

type PrivacyRepository interface {
	Get(context.Context, string) (domain.PrivacyPolicy, int64, error)
	Save(context.Context, domain.PrivacyPolicy, int64) (int64, error)
}

type ExportRepository interface {
	FindByIdempotency(context.Context, string, string) (domain.ExportRequest, domain.ExportResult, error)
	Create(context.Context, domain.ExportRequest, domain.ExportResult) error
	Get(context.Context, string) (domain.ExportRequest, domain.ExportResult, error)
}

type FeedbackRepository interface {
	Create(context.Context, domain.TemplateFeedback) error
	Get(context.Context, string) (domain.TemplateFeedback, error)
	List(context.Context, FeedbackListFilter) ([]domain.TemplateFeedback, int, error)
	Update(context.Context, domain.TemplateFeedback) error
}

type AttachmentRepository interface {
	Create(context.Context, domain.Attachment) error
	Get(context.Context, string) (domain.Attachment, error)
}

type AuditRepository interface {
	Append(context.Context, domain.AuditEvent) error
	List(context.Context, AuditListFilter) ([]domain.AuditEvent, int, error)
}

type TransactionManager interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}

type FileStore interface {
	Put(context.Context, string, io.Reader) (path string, sha256 string, size int64, err error)
	Open(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}

type Renderer interface {
	Render(context.Context, domain.TemplateVersion, domain.Draft, domain.PrivacyPolicy, domain.ExportFormat) ([]byte, []string, error)
}

type Notifier interface {
	Notify(context.Context, Notification) error
}

type Clock interface{ Now() time.Time }
type IDGenerator interface{ NewID(prefix string) string }

type Notification struct {
	RecipientID string
	Topic       string
	Body        string
}

type PageRequest struct {
	Page     int
	PageSize int
	Sort     string
	Order    string
}

func (p PageRequest) Normalize(allowedSort map[string]bool) PageRequest {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	if !allowedSort[p.Sort] {
		p.Sort = "updated_at"
	}
	if p.Order != "asc" {
		p.Order = "desc"
	}
	return p
}

type TemplateListFilter struct {
	PageRequest
	Status   domain.TemplateStatus
	Scenario domain.TargetScenario
	Category string
}
type DraftListFilter struct {
	PageRequest
	OwnerID  string
	Status   domain.DraftStatus
	TargetID string
}
type FeedbackListFilter struct {
	PageRequest
	Status            domain.FeedbackStatus
	TemplateVersionID string
}
type AuditListFilter struct {
	PageRequest
	ActorID, Resource, ResourceID, Action string
}
