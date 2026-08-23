package domain

import (
	"reflect"
	"testing"
	"time"
)

func TestSwitchTemplatePreservesMappedAndUnmappedValues(t *testing.T) {
	draft := Draft{ID: "d1", TemplateVersionID: "old", Values: map[string]any{"name": "林未", "community": "维护 Go 工具"}, Unmapped: map[string]any{"legacy": "保留"}, Version: 4}
	target := TemplateVersion{ID: "new", Sections: []TemplateSection{{Key: "base", Title: "资料", Fields: []FieldSpec{{Key: "full_name", Label: "姓名", Aliases: []string{"name"}, Required: true}}}}}
	result := SwitchTemplate(draft, target, TemplateMapping{FromTemplateVersionID: "old", ToTemplateVersionID: "new"}, ScenarioSocial, time.Now())
	if result.Draft.Values["full_name"] != "林未" {
		t.Fatalf("mapped value lost: %#v", result.Draft.Values)
	}
	if result.Draft.Unmapped["community"] != "维护 Go 工具" || result.Draft.Unmapped["legacy"] != "保留" {
		t.Fatalf("unmapped values lost: %#v", result.Draft.Unmapped)
	}
	if !reflect.DeepEqual(result.UnmappedFields, []string{"community"}) {
		t.Fatalf("unexpected unmapped list: %#v", result.UnmappedFields)
	}
}

func TestSnapshotDoesNotAliasDraftValues(t *testing.T) {
	draft := Draft{ID: "d1", Values: map[string]any{"skills": []any{"Go", "SQL"}}, Unmapped: map[string]any{}, Version: 2}
	snapshot := NewSnapshot("s1", draft, "manual", time.Now())
	draft.Values["skills"].([]any)[0] = "Rust"
	if snapshot.Values["skills"].([]any)[0] != "Go" {
		t.Fatal("snapshot changed through draft alias")
	}
}
