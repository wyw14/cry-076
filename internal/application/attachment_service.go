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
	decision := evaluateAttachmentAccess(actor, item, policy)
	if !decision.allowed {
		s.recordAttachmentDecision(ctx, actor, item, requestID, decision)
		return nil, domain.ErrForbidden
	}
	reader, err := s.files.Open(ctx, item.Path)
	if err != nil {
		decision.reason = "local_file_unavailable"
		s.recordAttachmentDecision(ctx, actor, item, requestID, decision)
		return nil, err
	}
	decision.reason = "content_opened"
	s.recordAttachmentDecision(ctx, actor, item, requestID, decision)
	return reader, nil
}

type attachmentAccessDecision struct {
	allowed bool
	reason  string
	source  string
}

func evaluateAttachmentAccess(actor domain.Actor, item domain.Attachment, policy domain.PrivacyPolicy) attachmentAccessDecision {
	if actor.CanManage(item.OwnerID) {
		return attachmentAccessDecision{allowed: true, reason: "owner_or_admin", source: "role"}
	}
	if actor.ID == "" {
		return attachmentAccessDecision{reason: "missing_actor", source: "identity"}
	}
	if policy.OwnerID != "" && policy.OwnerID != item.OwnerID {
		return attachmentAccessDecision{reason: "policy_owner_mismatch", source: "privacy_policy"}
	}
	if policy.AllowedAttachments == nil {
		return attachmentAccessDecision{
			allowed: actor.CanReview(),
			reason:  "legacy_policy_without_attachment_rules",
			source:  "privacy_policy",
		}
	}
	allowed, declared := policy.AllowedAttachments[item.ID]
	if !declared {
		return attachmentAccessDecision{reason: "attachment_not_declared", source: "privacy_policy"}
	}
	if !allowed {
		return attachmentAccessDecision{reason: "attachment_explicitly_blocked", source: "privacy_policy"}
	}
	return attachmentAccessDecision{allowed: true, reason: "attachment_explicitly_shared", source: "privacy_policy"}
}

func (s *AttachmentService) recordAttachmentDecision(ctx context.Context, actor domain.Actor, item domain.Attachment, requestID string, decision attachmentAccessDecision) {
	outcome := "denied"
	action := "attachment.open_denied"
	if decision.allowed {
		outcome = "allowed"
		action = "attachment.opened"
	}
	_ = s.audits.Append(ctx, domain.AuditEvent{
		ID: s.ids.NewID("audit"), ActorID: actor.ID,
		Action: action, Resource: "attachment", ResourceID: item.ID,
		RequestID: requestID, Outcome: outcome,
		Metadata: map[string]any{
			"decision_reason": decision.reason,
			"decision_source": decision.source,
			"owner_id":        item.OwnerID,
		},
		CreatedAt: s.clock.Now(),
	})
}
