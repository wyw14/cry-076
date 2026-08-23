package domain

import "testing"

func TestMaterialConsistencySeparatesMissingPrivateAndUnmappedFields(t *testing.T) {
	template := TemplateVersion{
		ID: "tv-technical-4", TemplateID: "technical", Version: 4, Status: TemplatePublished,
		Scenarios: []TargetScenario{ScenarioTechnical},
		Sections: []TemplateSection{{Key: "identity", Fields: []FieldSpec{
			{Key: "full_name", Required: true},
			{Key: "email", Required: true, Private: true},
			{Key: "portfolio", Required: true, Visibility: VisibilityRule{Requires: []string{"full_name"}}},
		}}},
	}
	draft := Draft{
		ID: "draft-7", TemplateVersionID: template.ID, Version: 9, Status: DraftActive,
		Values:   map[string]any{"full_name": "林未", "email": "lin@example.test", "legacy_summary": "保留内容"},
		Unmapped: map[string]any{"old_skill": "COBOL"},
	}
	policy := PrivacyPolicy{OwnerID: "owner-1", IncludedFields: map[string]bool{"email": false}}
	report := EvaluateMaterialConsistency(template, draft, policy, ScenarioTechnical)
	if report.PreviewReady || report.ExportReady {
		t.Fatal("missing visible portfolio must block preview and export")
	}
	if len(report.PrivateOmissions) != 1 || report.PrivateOmissions[0] != "email" {
		t.Fatalf("private omissions=%v", report.PrivateOmissions)
	}
	if len(report.UnmappedFields) != 2 || report.UnmappedFields[0] != "legacy_summary" || report.UnmappedFields[1] != "old_skill" {
		t.Fatalf("unmapped=%v", report.UnmappedFields)
	}
	if len(report.ConsistencyDigest) != 64 {
		t.Fatalf("digest=%q", report.ConsistencyDigest)
	}
}

func TestMaterialConsistencyAllowsDeprecatedTemplateForExistingDraft(t *testing.T) {
	template := TemplateVersion{ID: "tv-old", Status: TemplateDeprecated, Scenarios: []TargetScenario{ScenarioSocial}, Sections: []TemplateSection{{Fields: []FieldSpec{{Key: "name", Required: true}}}}}
	draft := Draft{ID: "draft-old", TemplateVersionID: template.ID, Version: 2, Status: DraftActive, Values: map[string]any{"name": "林未"}}
	report := EvaluateMaterialConsistency(template, draft, PrivacyPolicy{IncludedFields: map[string]bool{}}, ScenarioSocial)
	if !report.PreviewReady || !report.ExportReady {
		t.Fatalf("existing material should remain usable after deprecation: %#v", report.Issues)
	}
}
