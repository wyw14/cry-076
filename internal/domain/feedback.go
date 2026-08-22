package domain

import "time"

type FeedbackStatus string

const (
	FeedbackOpen      FeedbackStatus = "open"
	FeedbackTriaged   FeedbackStatus = "triaged"
	FeedbackResolved  FeedbackStatus = "resolved"
	FeedbackDismissed FeedbackStatus = "dismissed"
)

type TemplateFeedback struct {
	ID                string         `json:"id"`
	TemplateVersionID string         `json:"template_version_id"`
	ReporterID        string         `json:"reporter_id"`
	Kind              string         `json:"kind"`
	Message           string         `json:"message"`
	Status            FeedbackStatus `json:"status"`
	Resolution        string         `json:"resolution,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

func (f TemplateFeedback) Transition(next FeedbackStatus, resolution string, now time.Time) (TemplateFeedback, error) {
	allowed := map[FeedbackStatus]map[FeedbackStatus]bool{
		FeedbackOpen:    {FeedbackTriaged: true, FeedbackDismissed: true},
		FeedbackTriaged: {FeedbackResolved: true, FeedbackDismissed: true},
	}
	if !allowed[f.Status][next] {
		return TemplateFeedback{}, ErrInvalidTransition
	}
	if (next == FeedbackResolved || next == FeedbackDismissed) && resolution == "" {
		return TemplateFeedback{}, ErrValidation
	}
	f.Status, f.Resolution, f.UpdatedAt = next, resolution, now
	return f, nil
}
