package domain

import "testing"

func TestVisibilityRequiresEveryDependency(t *testing.T) {
	rule := VisibilityRule{
		Scenarios: []TargetScenario{ScenarioTechnical},
		Requires:  []string{"full_name", "portfolio_url"},
	}
	values := map[string]any{"full_name": "林未"}
	if rule.Visible(ScenarioTechnical, values) {
		t.Fatal("field must stay hidden until every prerequisite is filled")
	}
	values["portfolio_url"] = "https://portfolio.example.test"
	if !rule.Visible(ScenarioTechnical, values) {
		t.Fatal("field should become visible after every prerequisite is filled")
	}
}

