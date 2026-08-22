package application_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
	"github.com/wyw14/cry-076/internal/platform"
)

type ids struct{ n int }

func (i *ids) NewID(prefix string) string { i.n++; return prefix + "-test-" + string(rune('a'+i.n)) }

type fixedClock struct{ Value time.Time }

func (f fixedClock) Now() time.Time { return f.Value }

type fixture struct {
	store       *fakeCore
	snapshots   fakeSnapshots
	mappings    fakeMappings
	privacy     fakePrivacy
	exports     fakeExports
	attachments fakeAttachments
	audits      *fakeAudits
	feedback    fakeFeedback
	clock       fixedClock
	ids         *ids
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	store := newFakeCore()
	return &fixture{store: store, snapshots: newFakeSnapshots(), mappings: newFakeMappings(), privacy: newFakePrivacy(), exports: newFakeExports(), attachments: newFakeAttachments(), audits: &fakeAudits{}, feedback: newFakeFeedback(), clock: fixedClock{Value: time.Date(2026, 8, 22, 1, 2, 3, 0, time.UTC)}, ids: &ids{}}
}
func publishedTemplate() domain.TemplateVersion {
	return domain.TemplateVersion{ID: "tv1", TemplateID: "template1", Version: 1, Name: "清晰双栏", Category: "通用", Scenarios: []domain.TargetScenario{domain.ScenarioSocial}, Status: domain.TemplatePublished, Sections: []domain.TemplateSection{{Key: "base", Title: "资料", Fields: []domain.FieldSpec{{Key: "full_name", Label: "姓名", Required: true}, {Key: "email", Label: "邮箱", Private: true}, {Key: "phone", Label: "电话", Private: true}}}}, Style: domain.PreviewStyle{FontFamily: "sans-serif", AccentColor: "#0f766e"}}
}
func seedDraft(t *testing.T, f *fixture) domain.Draft {
	t.Helper()
	template := publishedTemplate()
	if err := f.store.CreateVersion(context.Background(), template); err != nil {
		t.Fatal(err)
	}
	draft := domain.Draft{ID: "d1", OwnerID: "u1", TargetID: "target1", TemplateVersionID: template.ID, Values: map[string]any{"full_name": "林未", "email": "lin@example.test", "phone": "13800000000"}, Unmapped: map[string]any{}, Version: 1, Status: domain.DraftActive, UpdatedAt: f.clock.Now()}
	created, err := f.store.Create(context.Background(), draft, "create-key", "create-hash")
	if err != nil {
		t.Fatal(err)
	}
	return created
}

func TestDraftAutosaveRejectsStaleVersion(t *testing.T) {
	f := newFixture(t)
	seedDraft(t, f)
	service := application.NewDraftService(f.store, f.snapshots, f.store, f.audits, f.store, f.clock, f.ids)
	actor := domain.Actor{ID: "u1", Role: domain.RoleOwner}
	saved, err := service.AutoSave(context.Background(), actor, application.SaveDraftInput{DraftID: "d1", ExpectedVersion: 1, Values: map[string]any{"full_name": "第一次"}, IdempotencyKey: "save1", RequestHash: "hash1"})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Version != 2 {
		t.Fatalf("version=%d", saved.Version)
	}
	_, err = service.AutoSave(context.Background(), actor, application.SaveDraftInput{DraftID: "d1", ExpectedVersion: 1, Values: map[string]any{"full_name": "过期写入"}, IdempotencyKey: "save2", RequestHash: "hash2"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestSwitchServiceCreatesRollbackSnapshot(t *testing.T) {
	f := newFixture(t)
	seedDraft(t, f)
	target := publishedTemplate()
	target.ID = "tv2"
	target.Sections[0].Fields = []domain.FieldSpec{{Key: "name", Label: "姓名", Required: true}, {Key: "portfolio", Label: "作品"}}
	if err := f.store.CreateVersion(context.Background(), target); err != nil {
		t.Fatal(err)
	}
	if err := f.mappings.Put(context.Background(), domain.TemplateMapping{FromTemplateVersionID: "tv1", ToTemplateVersionID: "tv2", Rules: []domain.MappingRule{{Source: "full_name", Target: "name"}}}); err != nil {
		t.Fatal(err)
	}
	service := application.NewSwitchService(f.store, f.store, f.mappings, f.snapshots, f.audits, f.store, f.clock, f.ids)
	result, err := service.Switch(context.Background(), domain.Actor{ID: "u1", Role: domain.RoleOwner}, application.SwitchTemplateInput{DraftID: "d1", TargetTemplateVersionID: "tv2", Scenario: domain.ScenarioSocial, ExpectedVersion: 1, IdempotencyKey: "switch", RequestHash: "switch-hash"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Draft.Values["name"] != "林未" {
		t.Fatalf("mapped value missing: %#v", result.Draft.Values)
	}
	snapshots, total, err := f.snapshots.ListByDraft(context.Background(), "d1", application.PageRequest{Page: 1, PageSize: 10})
	if err != nil || total != 1 || snapshots[0].Reason != "before_template_switch" {
		t.Fatalf("rollback snapshot missing: total=%d err=%v", total, err)
	}
}

func TestExportContainsOnlySelectedFields(t *testing.T) {
	f := newFixture(t)
	seedDraft(t, f)
	if _, err := f.privacy.Save(context.Background(), domain.PrivacyPolicy{OwnerID: "u1", IncludedFields: map[string]bool{"full_name": true, "email": false, "phone": false}, MaskInPreview: map[string]bool{}, AllowedAttachments: map[string]bool{}}, 0); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	service := application.NewExportService(f.store, f.snapshots, f.store, f.privacy, f.exports, platform.LocalFileStore{Root: root}, platform.OfflineRenderer{}, f.audits, f.clock, f.ids)
	result, err := service.Create(context.Background(), domain.Actor{ID: "u1", Role: domain.RoleOwner}, application.CreateExportInput{DraftID: "d1", Format: domain.ExportJSON, IdempotencyKey: "export1", RequestHash: "export-hash"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.FieldKeys) != 1 || result.FieldKeys[0] != "full_name" {
		t.Fatalf("privacy leak in fields: %#v", result.FieldKeys)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(result.Path))); err != nil {
		t.Fatalf("export missing: %v", err)
	}
	validation, err := service.Validate(context.Background(), domain.Actor{ID: "u1", Role: domain.RoleOwner}, result.ID)
	if err != nil || !validation.Valid {
		t.Fatalf("validation=%#v err=%v", validation, err)
	}
}

func TestAttachmentAccessUsesPolicyAndAudit(t *testing.T) {
	f := newFixture(t)
	root := t.TempDir()
	service := application.NewAttachmentService(f.attachments, f.privacy, platform.LocalFileStore{Root: root}, f.audits, f.clock, f.ids, 1024, []string{"application/pdf"})
	item, err := service.Upload(context.Background(), domain.Actor{ID: "u1", Role: domain.RoleOwner}, "u1", "certificate.pdf", "application/pdf", "r1", strings.NewReader("pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.privacy.Save(context.Background(), domain.PrivacyPolicy{OwnerID: "u1", IncludedFields: map[string]bool{}, AllowedAttachments: map[string]bool{item.ID: true}}, 0); err != nil {
		t.Fatal(err)
	}
	reader, err := service.Open(context.Background(), domain.Actor{ID: "reviewer", Role: domain.RoleReviewer}, item.ID, "r2")
	if err != nil {
		t.Fatal(err)
	}
	reader.Close()
	events, total, err := f.audits.List(context.Background(), application.AuditListFilter{PageRequest: application.PageRequest{Page: 1, PageSize: 10}, Resource: "attachment"})
	if err != nil || total != 2 || len(events) != 2 {
		t.Fatalf("audit total=%d events=%d err=%v", total, len(events), err)
	}
}

func TestFeedbackRequiresReviewTransitions(t *testing.T) {
	f := newFixture(t)
	template := publishedTemplate()
	if err := f.store.CreateVersion(context.Background(), template); err != nil {
		t.Fatal(err)
	}
	notifier := &platform.ResumeNotifier{}
	service := application.NewFeedbackService(f.feedback, f.store, f.audits, notifier, f.clock, f.ids)
	item, err := service.Submit(context.Background(), domain.Actor{ID: "u1", Role: domain.RoleOwner}, template.ID, "layout", "标题间距不稳定", "r1")
	if err != nil {
		t.Fatal(err)
	}
	triaged, err := service.Review(context.Background(), domain.Actor{ID: "reviewer", Role: domain.RoleReviewer}, item.ID, domain.FeedbackTriaged, "已确认", "r2")
	if err != nil {
		t.Fatal(err)
	}
	if triaged.Status != domain.FeedbackTriaged || len(notifier.Messages) != 1 {
		t.Fatalf("triaged=%s notifications=%d", triaged.Status, len(notifier.Messages))
	}
}

func TestTemplatePublicationBlocksUnresolvedFeedback(t *testing.T) {
	f := newFixture(t)
	template := publishedTemplate()
	template.Status = domain.TemplateReview
	if err := f.store.CreateVersion(context.Background(), template); err != nil {
		t.Fatal(err)
	}
	if err := f.feedback.Create(context.Background(), domain.TemplateFeedback{ID: "blocking", TemplateVersionID: template.ID, ReporterID: "u1", Kind: "blocking", Message: "导出样式错位", Status: domain.FeedbackOpen, CreatedAt: f.clock.Now(), UpdatedAt: f.clock.Now()}); err != nil {
		t.Fatal(err)
	}
	service := application.NewTemplateService(f.store, f.feedback, f.audits, &platform.ResumeNotifier{}, f.clock, f.ids)
	_, err := service.Transition(context.Background(), domain.Actor{ID: "reviewer", Role: domain.RoleReviewer}, template.ID, template.Version, domain.TemplatePublished, "r1")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected blocking feedback conflict, got %v", err)
	}
}
