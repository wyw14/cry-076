package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/wyw14/cry-076/internal/domain"
)

type ExportService struct {
	drafts    DraftRepository
	snapshots SnapshotRepository
	templates TemplateRepository
	privacy   PrivacyRepository
	exports   ExportRepository
	files     FileStore
	renderer  Renderer
	audits    AuditRepository
	clock     Clock
	ids       IDGenerator
}

func NewExportService(d DraftRepository, s SnapshotRepository, t TemplateRepository, p PrivacyRepository, e ExportRepository, f FileStore, r Renderer, a AuditRepository, c Clock, ids IDGenerator) *ExportService {
	return &ExportService{drafts: d, snapshots: s, templates: t, privacy: p, exports: e, files: f, renderer: r, audits: a, clock: c, ids: ids}
}

type CreateExportInput struct {
	DraftID, SnapshotID                    string
	Format                                 domain.ExportFormat
	IdempotencyKey, RequestHash, RequestID string
}

func (s *ExportService) Create(ctx context.Context, actor domain.Actor, input CreateExportInput) (domain.ExportResult, error) {
	if existingReq, result, err := s.exports.FindByIdempotency(ctx, actor.ID, input.IdempotencyKey); err == nil {
		if existingReq.IdempotencyKey != input.IdempotencyKey || existingReq.DraftID != input.DraftID || existingReq.SnapshotID != input.SnapshotID || existingReq.Format != input.Format {
			return domain.ExportResult{}, domain.ErrIdempotencyReuse
		}
		return result, nil
	} else if err != domain.ErrNotFound {
		return domain.ExportResult{}, err
	}
	draft, err := s.drafts.Get(ctx, input.DraftID)
	if err != nil {
		return domain.ExportResult{}, err
	}
	if !actor.CanManage(draft.OwnerID) {
		return domain.ExportResult{}, domain.ErrForbidden
	}
	if input.SnapshotID != "" {
		snapshot, err := s.snapshots.Get(ctx, input.SnapshotID)
		if err != nil {
			return domain.ExportResult{}, err
		}
		if snapshot.DraftID != draft.ID {
			return domain.ExportResult{}, domain.ErrValidation
		}
		draft.TemplateVersionID, draft.Values, draft.Unmapped = snapshot.TemplateVersionID, snapshot.Values, snapshot.Unmapped
	}
	template, err := s.templates.GetVersion(ctx, draft.TemplateVersionID)
	if err != nil {
		return domain.ExportResult{}, err
	}
	policy, privacyVersion, err := s.privacy.Get(ctx, draft.OwnerID)
	if err != nil {
		return domain.ExportResult{}, err
	}
	content, keys, err := s.renderer.Render(ctx, template, draft, policy, input.Format)
	if err != nil {
		return domain.ExportResult{}, fmt.Errorf("render export: %w", err)
	}
	for _, key := range keys {
		if !policy.Includes(key) {
			return domain.ExportResult{}, fmt.Errorf("%w: renderer included private field %s", domain.ErrValidation, key)
		}
	}
	path, hash, size, err := s.files.Put(ctx, "exports/"+s.ids.NewID("export"), bytes.NewReader(content))
	if err != nil {
		return domain.ExportResult{}, err
	}
	request := domain.ExportRequest{ID: s.ids.NewID("export_request"), OwnerID: draft.OwnerID, DraftID: draft.ID, SnapshotID: input.SnapshotID, Format: input.Format, PrivacyVersion: privacyVersion, IdempotencyKey: input.IdempotencyKey}
	sort.Strings(keys)
	result := domain.ExportResult{ID: s.ids.NewID("export"), RequestID: request.ID, Format: input.Format, Path: path, SHA256: hash, Size: size, FieldKeys: keys, CreatedAt: s.clock.Now()}
	if err := s.exports.Create(ctx, request, result); err != nil {
		_ = s.files.Delete(ctx, path)
		return domain.ExportResult{}, err
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "export.created", Resource: "export", ResourceID: result.ID, RequestID: input.RequestID, Metadata: map[string]any{"fields": keys, "privacy_version": privacyVersion}, CreatedAt: s.clock.Now()})
	return result, nil
}

func (s *ExportService) Validate(ctx context.Context, actor domain.Actor, exportID string) (domain.ExportValidation, error) {
	request, result, err := s.exports.Get(ctx, exportID)
	if err != nil {
		return domain.ExportValidation{}, err
	}
	if !actor.CanManage(request.OwnerID) {
		return domain.ExportValidation{}, domain.ErrForbidden
	}
	reader, err := s.files.Open(ctx, result.Path)
	if err != nil {
		return domain.ExportValidation{}, err
	}
	defer reader.Close()
	content := new(bytes.Buffer)
	if _, err := content.ReadFrom(reader); err != nil {
		return domain.ExportValidation{}, err
	}
	sum := sha256.Sum256(content.Bytes())
	validation := domain.ExportValidation{Valid: true, HashMatches: hex.EncodeToString(sum[:]) == result.SHA256}
	policy, _, err := s.privacy.Get(ctx, request.OwnerID)
	if err != nil {
		return domain.ExportValidation{}, err
	}
	for _, key := range result.FieldKeys {
		if !policy.Includes(key) {
			validation.Unexpected = append(validation.Unexpected, key)
		}
	}
	validation.Valid = validation.HashMatches && len(validation.Unexpected) == 0
	return validation, nil
}
