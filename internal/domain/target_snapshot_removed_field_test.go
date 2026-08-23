package domain

import "testing"

func TestSnapshotComparisonReportsRemovedField(t *testing.T) {
	left := DraftSnapshot{Version: 8, Values: map[string]any{
		"full_name": "林未", "portfolio": "离线简历生成器",
	}}
	right := DraftSnapshot{Version: 9, Values: map[string]any{"full_name": "林未"}}
	comparison := CompareSnapshots(left, right)
	if len(comparison.Changes) != 1 {
		t.Fatalf("removed field must be visible in history comparison, got %#v", comparison.Changes)
	}
	change := comparison.Changes[0]
	if change.Field != "portfolio" || change.Before != "离线简历生成器" || change.After != nil {
		t.Fatalf("unexpected removed-field change: %#v", change)
	}
}

