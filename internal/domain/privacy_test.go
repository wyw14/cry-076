package domain

import "testing"

func TestPrivacyMaskAndSelection(t *testing.T) {
	policy := PrivacyPolicy{IncludedFields: map[string]bool{"full_name": true, "phone": false, "email": true}, MaskInPreview: map[string]bool{"email": true}}
	if got := policy.Mask("email", "a@example.test"); got != "a************t" {
		t.Fatalf("unexpected mask %q", got)
	}
	selected := policy.SelectedFields()
	if len(selected) != 2 || selected[0] != "email" || selected[1] != "full_name" {
		t.Fatalf("unexpected selected fields %#v", selected)
	}
}
