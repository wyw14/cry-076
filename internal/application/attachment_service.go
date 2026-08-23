package application

import (
	"context"
	"fmt"
	"io"

	"github.com/wyw14/cry-076/internal/domain"
)

type AttachmentService struct {
	attachments AttachmentRepository
	privacy     PrivacyRepository
	files       FileStore
	audits      AuditRepository
	clock       Clock
	ids         IDGenerator
	maxSize     int64
	allowed     map[string]bool
}

func NewAttachmentService(r AttachmentRepository, p PrivacyRepository, f FileStore, a AuditRepository, c Clock, ids IDGenerator, maxSize int64, allowed []string) *AttachmentService {
	types := make(map[string]bool, len(allowed))
	for _, item := range allowed {
		types[item] = true
	}
	return &AttachmentService{attachments: r, privacy: p, files: f, audits: a, clock: c, ids: ids, maxSize: maxSize, allowed: types}
}

func (s *AttachmentService) Upload(ctx context.Context, actor domain.Actor, ownerID, name, mediaType, requestID string, body io.Reader) (domain.Attachment, error) {
	if !actor.CanManage(ownerID) {
		return domain.Attachment{}, domain.ErrForbidden
	}
	id := s.ids.NewID("attachment")
	path, hash, size, err := s.files.Put(ctx, "attachments/"+id, io.LimitReader(body, s.maxSize+1))
	if err != nil {
		return domain.Attachment{}, err
	}
	item := domain.Attachment{ID: id, OwnerID: ownerID, Name: name, MediaType: mediaType, Size: size, SHA256: hash, Path: path, CreatedAt: s.clock.Now()}
	if err := item.Validate(s.maxSize, s.allowed); err != nil {
		_ = s.files.Delete(ctx, path)
		return domain.Attachment{}, err
	}
	if err := s.attachments.Create(ctx, item); err != nil {
		_ = s.files.Delete(ctx, path)
		return domain.Attachment{}, fmt.Errorf("create attachment: %w", err)
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "attachment.uploaded", Resource: "attachment", ResourceID: id, RequestID: requestID, CreatedAt: s.clock.Now()})
	return item, nil
}

func (s *AttachmentService) Open(ctx context.Context, actor domain.Actor, id, requestID string) (io.ReadCloser, error) {
	item, err := s.attachments.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	policy, _, err := s.privacy.Get(ctx, item.OwnerID)
	if err != nil {
		return nil, err
	}
	if !actor.CanManage(item.OwnerID) && !policy.AllowedAttachments[id] {
		return nil, domain.ErrForbidden
	}
	reader, err := s.files.Open(ctx, item.Path)
	if err != nil {
		return nil, err
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{ID: s.ids.NewID("audit"), ActorID: actor.ID, Action: "attachment.opened", Resource: "attachment", ResourceID: id, RequestID: requestID, CreatedAt: s.clock.Now()})
	return reader, nil
}
