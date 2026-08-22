package application_test

import (
	"context"
	"sort"
	"time"

	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
)

type fakeCore struct {
	templates map[string]domain.TemplateVersion
	drafts    map[string]domain.Draft
	draftKeys map[string]struct {
		hash  string
		value domain.Draft
	}
}

func newFakeCore() *fakeCore {
	return &fakeCore{templates: map[string]domain.TemplateVersion{}, drafts: map[string]domain.Draft{}, draftKeys: map[string]struct {
		hash  string
		value domain.Draft
	}{}}
}
func (f *fakeCore) WithinTransaction(ctx context.Context, operation func(context.Context) error) error {
	return operation(ctx)
}
func (f *fakeCore) CreateVersion(_ context.Context, item domain.TemplateVersion) error {
	if _, ok := f.templates[item.ID]; ok {
		return domain.ErrConflict
	}
	f.templates[item.ID] = item
	return nil
}
func (f *fakeCore) GetVersion(_ context.Context, id string) (domain.TemplateVersion, error) {
	item, ok := f.templates[id]
	if !ok {
		return domain.TemplateVersion{}, domain.ErrNotFound
	}
	return item, nil
}
func (f *fakeCore) ListVersions(_ context.Context, filter application.TemplateListFilter) ([]domain.TemplateVersion, int, error) {
	items := make([]domain.TemplateVersion, 0)
	for _, item := range f.templates {
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		if filter.Scenario != "" && !item.Supports(filter.Scenario) {
			continue
		}
		items = append(items, item)
	}
	return items, len(items), nil
}
func (f *fakeCore) UpdateVersion(_ context.Context, item domain.TemplateVersion, expected int64) error {
	current, ok := f.templates[item.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if current.Version != expected {
		return domain.ErrConflict
	}
	f.templates[item.ID] = item
	return nil
}
func (f *fakeCore) Create(_ context.Context, item domain.Draft, key, hash string) (domain.Draft, error) {
	lookup := item.OwnerID + ":" + key
	if existing, ok := f.draftKeys[lookup]; ok {
		if existing.hash != hash {
			return domain.Draft{}, domain.ErrIdempotencyReuse
		}
		return existing.value.Clone(), nil
	}
	f.drafts[item.ID] = item.Clone()
	f.draftKeys[lookup] = struct {
		hash  string
		value domain.Draft
	}{hash, item.Clone()}
	return item.Clone(), nil
}
func (f *fakeCore) Get(_ context.Context, id string) (domain.Draft, error) {
	item, ok := f.drafts[id]
	if !ok {
		return domain.Draft{}, domain.ErrNotFound
	}
	return item.Clone(), nil
}
func (f *fakeCore) List(_ context.Context, filter application.DraftListFilter) ([]domain.Draft, int, error) {
	items := make([]domain.Draft, 0)
	for _, item := range f.drafts {
		if filter.OwnerID != "" && item.OwnerID != filter.OwnerID {
			continue
		}
		items = append(items, item.Clone())
	}
	return items, len(items), nil
}
func (f *fakeCore) Save(_ context.Context, item domain.Draft, expected int64, key, hash string) (domain.Draft, error) {
	lookup := item.OwnerID + ":" + key
	if existing, ok := f.draftKeys[lookup]; ok {
		if existing.hash != hash {
			return domain.Draft{}, domain.ErrIdempotencyReuse
		}
		return existing.value.Clone(), nil
	}
	current, ok := f.drafts[item.ID]
	if !ok {
		return domain.Draft{}, domain.ErrNotFound
	}
	if current.Version != expected {
		return domain.Draft{}, domain.ErrConflict
	}
	f.drafts[item.ID] = item.Clone()
	f.draftKeys[lookup] = struct {
		hash  string
		value domain.Draft
	}{hash, item.Clone()}
	return item.Clone(), nil
}
func (f *fakeCore) Archive(_ context.Context, id string, expected int64, now time.Time) error {
	item, ok := f.drafts[id]
	if !ok {
		return domain.ErrNotFound
	}
	if item.Version != expected {
		return domain.ErrConflict
	}
	item.Status = domain.DraftArchived
	item.Version++
	item.UpdatedAt = now
	f.drafts[id] = item
	return nil
}

type fakeSnapshots struct {
	items map[string]domain.DraftSnapshot
}

func newFakeSnapshots() fakeSnapshots { return fakeSnapshots{items: map[string]domain.DraftSnapshot{}} }
func (f fakeSnapshots) Create(_ context.Context, item domain.DraftSnapshot) error {
	f.items[item.ID] = item
	return nil
}
func (f fakeSnapshots) Get(_ context.Context, id string) (domain.DraftSnapshot, error) {
	item, ok := f.items[id]
	if !ok {
		return domain.DraftSnapshot{}, domain.ErrNotFound
	}
	return item, nil
}
func (f fakeSnapshots) ListByDraft(_ context.Context, id string, _ application.PageRequest) ([]domain.DraftSnapshot, int, error) {
	items := make([]domain.DraftSnapshot, 0)
	for _, item := range f.items {
		if item.DraftID == id {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, len(items), nil
}

type fakeMappings struct {
	items map[string]domain.TemplateMapping
}

func newFakeMappings() fakeMappings { return fakeMappings{items: map[string]domain.TemplateMapping{}} }
func (f fakeMappings) Get(_ context.Context, from, to string) (domain.TemplateMapping, error) {
	item, ok := f.items[from+":"+to]
	if !ok {
		return domain.TemplateMapping{}, domain.ErrNotFound
	}
	return item, nil
}
func (f fakeMappings) Put(_ context.Context, item domain.TemplateMapping) error {
	f.items[item.FromTemplateVersionID+":"+item.ToTemplateVersionID] = item
	return nil
}

type fakePrivacy struct {
	items    map[string]domain.PrivacyPolicy
	versions map[string]int64
}

func newFakePrivacy() fakePrivacy {
	return fakePrivacy{items: map[string]domain.PrivacyPolicy{}, versions: map[string]int64{}}
}
func (f fakePrivacy) Get(_ context.Context, owner string) (domain.PrivacyPolicy, int64, error) {
	item, ok := f.items[owner]
	if !ok {
		return domain.PrivacyPolicy{}, 0, domain.ErrNotFound
	}
	return item, f.versions[owner], nil
}
func (f fakePrivacy) Save(_ context.Context, item domain.PrivacyPolicy, expected int64) (int64, error) {
	if f.versions[item.OwnerID] != expected {
		return 0, domain.ErrConflict
	}
	f.versions[item.OwnerID]++
	f.items[item.OwnerID] = item
	return f.versions[item.OwnerID], nil
}

type fakeExports struct {
	items map[string]struct {
		request domain.ExportRequest
		result  domain.ExportResult
	}
	keys map[string]string
}

func newFakeExports() fakeExports {
	return fakeExports{items: map[string]struct {
		request domain.ExportRequest
		result  domain.ExportResult
	}{}, keys: map[string]string{}}
}
func (f fakeExports) FindByIdempotency(_ context.Context, owner, key string) (domain.ExportRequest, domain.ExportResult, error) {
	id, ok := f.keys[owner+":"+key]
	if !ok {
		return domain.ExportRequest{}, domain.ExportResult{}, domain.ErrNotFound
	}
	item := f.items[id]
	return item.request, item.result, nil
}
func (f fakeExports) Create(_ context.Context, request domain.ExportRequest, result domain.ExportResult) error {
	f.items[result.ID] = struct {
		request domain.ExportRequest
		result  domain.ExportResult
	}{request, result}
	f.keys[request.OwnerID+":"+request.IdempotencyKey] = result.ID
	return nil
}
func (f fakeExports) Get(_ context.Context, id string) (domain.ExportRequest, domain.ExportResult, error) {
	item, ok := f.items[id]
	if !ok {
		return domain.ExportRequest{}, domain.ExportResult{}, domain.ErrNotFound
	}
	return item.request, item.result, nil
}

type fakeAttachments struct{ items map[string]domain.Attachment }

func newFakeAttachments() fakeAttachments {
	return fakeAttachments{items: map[string]domain.Attachment{}}
}
func (f fakeAttachments) Create(_ context.Context, item domain.Attachment) error {
	f.items[item.ID] = item
	return nil
}
func (f fakeAttachments) Get(_ context.Context, id string) (domain.Attachment, error) {
	item, ok := f.items[id]
	if !ok {
		return domain.Attachment{}, domain.ErrNotFound
	}
	return item, nil
}

type fakeAudits struct{ items []domain.AuditEvent }

func (f *fakeAudits) Append(_ context.Context, item domain.AuditEvent) error {
	f.items = append(f.items, item)
	return nil
}
func (f *fakeAudits) List(_ context.Context, filter application.AuditListFilter) ([]domain.AuditEvent, int, error) {
	items := make([]domain.AuditEvent, 0)
	for _, item := range f.items {
		if filter.Resource != "" && item.Resource != filter.Resource {
			continue
		}
		items = append(items, item)
	}
	return items, len(items), nil
}

type fakeFeedback struct {
	items map[string]domain.TemplateFeedback
}

func newFakeFeedback() fakeFeedback { return fakeFeedback{items: map[string]domain.TemplateFeedback{}} }
func (f fakeFeedback) Create(_ context.Context, item domain.TemplateFeedback) error {
	f.items[item.ID] = item
	return nil
}
func (f fakeFeedback) Get(_ context.Context, id string) (domain.TemplateFeedback, error) {
	item, ok := f.items[id]
	if !ok {
		return domain.TemplateFeedback{}, domain.ErrNotFound
	}
	return item, nil
}
func (f fakeFeedback) List(_ context.Context, filter application.FeedbackListFilter) ([]domain.TemplateFeedback, int, error) {
	items := make([]domain.TemplateFeedback, 0)
	for _, item := range f.items {
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		if filter.TemplateVersionID != "" && item.TemplateVersionID != filter.TemplateVersionID {
			continue
		}
		items = append(items, item)
	}
	return items, len(items), nil
}
func (f fakeFeedback) Update(_ context.Context, item domain.TemplateFeedback) error {
	f.items[item.ID] = item
	return nil
}
