package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
	"github.com/wyw14/cry-076/internal/platform"
)

type failingFeedbackRepository struct{ fakeFeedback }

func (f failingFeedbackRepository) Update(context.Context, domain.TemplateFeedback) error {
	return errors.New("feedback write failed")
}

func TestFeedbackDoesNotNotifyWhenReviewWriteFails(t *testing.T) {
	fixture := newFixture(t)
	repository := failingFeedbackRepository{fakeFeedback: newFakeFeedback()}
	item := domain.TemplateFeedback{ID: "feedback", TemplateVersionID: "template", ReporterID: "reporter", Kind: "layout", Message: "spacing shifts", Status: domain.FeedbackTriaged, CreatedAt: fixture.clock.Now(), UpdatedAt: fixture.clock.Now()}
	if err := repository.Create(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	notifier := &platform.ResumeNotifier{}
	service := application.NewFeedbackService(repository, fixture.store, fixture.audits, notifier, fixture.clock, fixture.ids)
	_, err := service.Review(context.Background(), domain.Actor{ID: "reviewer", Role: domain.RoleReviewer}, item.ID, domain.FeedbackResolved, "resolved after review", "request-id")
	if err == nil {
		t.Fatal("feedback write unexpectedly succeeded")
	}
	if len(notifier.Messages) != 0 {
		t.Fatalf("success notification was sent before feedback persisted: %+v", notifier.Messages)
	}
}
