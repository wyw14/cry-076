package domain

import "strings"

type TargetScenario string

const (
	ScenarioCampus    TargetScenario = "campus"
	ScenarioSocial    TargetScenario = "social"
	ScenarioTechnical TargetScenario = "technical"
	ScenarioCreative  TargetScenario = "creative"
)

type JobTarget struct {
	ID            string         `json:"id"`
	OwnerID       string         `json:"owner_id"`
	Scenario      TargetScenario `json:"scenario"`
	Title         string         `json:"title"`
	Industry      string         `json:"industry"`
	Keywords      []string       `json:"keywords"`
	PreferredLang string         `json:"preferred_language"`
}

func (t JobTarget) Validate() error {
	if strings.TrimSpace(t.OwnerID) == "" || strings.TrimSpace(t.Title) == "" {
		return ErrValidation
	}
	switch t.Scenario {
	case ScenarioCampus, ScenarioSocial, ScenarioTechnical, ScenarioCreative:
		return nil
	default:
		return ErrValidation
	}
}
