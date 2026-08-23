package application_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
	"github.com/wyw14/cry-076/internal/platform"
)

type rejectingExports struct{}

func (rejectingExports) FindByIdempotency(context.Context, string, string) (domain.ExportRequest, domain.ExportResult, error) {
	return domain.ExportRequest{}, domain.ExportResult{}, domain.ErrNotFound
}
func (rejectingExports) Create(context.Context, domain.ExportRequest, domain.ExportResult) error {
	return errors.New("export repository unavailable")
}
func (rejectingExports) Get(context.Context, string) (domain.ExportRequest, domain.ExportResult, error) {
	return domain.ExportRequest{}, domain.ExportResult{}, domain.ErrNotFound
}

type trackingFiles struct {
	storedPath  string
	deletedPath string
}

func (f *trackingFiles) Put(_ context.Context, path string, body io.Reader) (string, string, int64, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return "", "", 0, err
	}
	f.storedPath = path
	return path, "hash", int64(len(data)), nil
}
func (f *trackingFiles) Open(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(nil)), nil
}
func (f *trackingFiles) Delete(_ context.Context, path string) error {
	f.deletedPath = path
	return nil
}

func TestExportRepositoryFailureRemovesStoredArtifact(t *testing.T) {
	f := newFixture(t)
	seedDraft(t, f)
	if _, err := f.privacy.Save(context.Background(), domain.PrivacyPolicy{OwnerID: "u1", IncludedFields: map[string]bool{"full_name": true}}, 0); err != nil {
		t.Fatal(err)
	}
	files := &trackingFiles{}
	service := application.NewExportService(f.store, f.snapshots, f.store, f.privacy, rejectingExports{}, files, platform.OfflineRenderer{}, f.audits, f.clock, f.ids)
	_, err := service.Create(context.Background(), domain.Actor{ID: "u1", Role: domain.RoleOwner}, application.CreateExportInput{DraftID: "d1", Format: domain.ExportJSON, IdempotencyKey: "cleanup", RequestHash: "cleanup-hash"})
	if err == nil {
		t.Fatal("expected repository failure")
	}
	if files.storedPath == "" {
		t.Fatal("export artifact was not stored before the repository failure")
	}
	if files.deletedPath != files.storedPath {
		t.Fatalf("deleted path=%q, want stored path %q", files.deletedPath, files.storedPath)
	}
}

