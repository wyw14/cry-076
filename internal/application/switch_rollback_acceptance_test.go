package application_test

import (
	"context"
	"testing"

	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
)

func TestTemplateSwitchCheckpointRestoresSourceTemplate(t *testing.T) {
	f := newFixture(t)
	seedDraft(t, f)
	target := publishedTemplate()
	target.ID = "tv2"
	target.Sections[0].Fields = []domain.FieldSpec{{Key: "name", Label: "姓名", Required: true}}
	if err := f.store.CreateVersion(context.Background(), target); err != nil {
		t.Fatal(err)
	}
	if err := f.mappings.Put(context.Background(), domain.TemplateMapping{FromTemplateVersionID: "tv1", ToTemplateVersionID: "tv2", Rules: []domain.MappingRule{{Source: "full_name", Target: "name"}}}); err != nil {
		t.Fatal(err)
	}
	service := application.NewSwitchService(f.store, f.store, f.mappings, f.snapshots, f.audits, f.store, f.clock, f.ids)
	if _, err := service.Switch(context.Background(), domain.Actor{ID: "u1", Role: domain.RoleOwner}, application.SwitchTemplateInput{DraftID: "d1", TargetTemplateVersionID: "tv2", Scenario: domain.ScenarioSocial, ExpectedVersion: 1, IdempotencyKey: "switch-rollback", RequestHash: "switch-rollback-hash"}); err != nil {
		t.Fatal(err)
	}
	snapshots, total, err := f.snapshots.ListByDraft(context.Background(), "d1", application.PageRequest{Page: 1, PageSize: 10})
	if err != nil || total != 1 {
		t.Fatalf("checkpoint total=%d err=%v", total, err)
	}
	checkpoint := snapshots[0]
	if checkpoint.TemplateVersionID != "tv1" {
		t.Fatalf("checkpoint captured template %q, want source tv1", checkpoint.TemplateVersionID)
	}
	if checkpoint.Values["full_name"] != "林未" {
		t.Fatalf("checkpoint lost source values: %#v", checkpoint.Values)
	}
}

