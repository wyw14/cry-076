package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
	"github.com/wyw14/cry-076/internal/platform"
)

func TestAttachmentShareDoesNotAuthorizeSiblingFile(t *testing.T) {
	f := newFixture(t)
	service := application.NewAttachmentService(f.attachments, f.privacy, platform.LocalFileStore{Root: t.TempDir()}, f.audits, f.clock, f.ids, 1024, []string{"application/pdf"})
	owner := domain.Actor{ID: "u1", Role: domain.RoleOwner}
	shared, err := service.Upload(context.Background(), owner, "u1", "shared.pdf", "application/pdf", "upload-1", strings.NewReader("shared"))
	if err != nil {
		t.Fatal(err)
	}
	private, err := service.Upload(context.Background(), owner, "u1", "private.pdf", "application/pdf", "upload-2", strings.NewReader("private"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.privacy.Save(context.Background(), domain.PrivacyPolicy{OwnerID: "u1", IncludedFields: map[string]bool{}, AllowedAttachments: map[string]bool{shared.ID: true}}, 0); err != nil {
		t.Fatal(err)
	}
	reviewer := domain.Actor{ID: "reviewer", Role: domain.RoleReviewer}
	reader, err := service.Open(context.Background(), reviewer, private.ID, "open-private")
	if reader != nil {
		reader.Close()
	}
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("opening unshared sibling returned %v, want forbidden", err)
	}
}

