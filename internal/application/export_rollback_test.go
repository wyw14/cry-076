package application_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
	"github.com/wyw14/cry-076/internal/platform"
)

// failingExports delegates to a real fakeExports but forces Create to fail,
// simulating a database outage while the local export file has already been
// written.
type failingExports struct{ inner fakeExports }

func (f failingExports) FindByIdempotency(ctx context.Context, owner, key string) (domain.ExportRequest, domain.ExportResult, error) {
	return f.inner.FindByIdempotency(ctx, owner, key)
}
func (failingExports) Create(_ context.Context, _ domain.ExportRequest, _ domain.ExportResult) error {
	return errors.New("database unavailable")
}
func (f failingExports) Get(ctx context.Context, id string) (domain.ExportRequest, domain.ExportResult, error) {
	return f.inner.Get(ctx, id)
}

// TestExportCreateRemovesOrphanFileWhenDatabaseCommitFails guards the rollback
// path: when the export record cannot be persisted, the locally written file
// must be deleted so retries do not accumulate unmanageable orphan files.
func TestExportCreateRemovesOrphanFileWhenDatabaseCommitFails(t *testing.T) {
	f := newFixture(t)
	seedDraft(t, f)
	if _, err := f.privacy.Save(context.Background(), domain.PrivacyPolicy{OwnerID: "u1", IncludedFields: map[string]bool{"full_name": true, "email": false, "phone": false}, MaskInPreview: map[string]bool{}, AllowedAttachments: map[string]bool{}}, 0); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	service := application.NewExportService(f.store, f.snapshots, f.store, f.privacy, failingExports{inner: f.exports}, platform.LocalFileStore{Root: root}, platform.OfflineRenderer{}, f.audits, f.clock, f.ids)

	_, err := service.Create(context.Background(), domain.Actor{ID: "u1", Role: domain.RoleOwner}, application.CreateExportInput{DraftID: "d1", Format: domain.ExportJSON, IdempotencyKey: "export1", RequestHash: "export-hash"})
	if err == nil {
		t.Fatal("expected database failure to surface as error")
	}

	// The exports directory must be empty: the written file was rolled back.
	entries, err := os.ReadDir(filepath.Join(root, "exports"))
	if err != nil {
		// Directory may not exist if nothing was written, which is also fine.
		return
	}
	for _, e := range entries {
		t.Fatalf("orphan file left behind after failed rollback: %s", e.Name())
	}
}
