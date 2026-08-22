package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
)

type MappingRepository struct{ DB *Database }

func (r MappingRepository) Get(ctx context.Context, from, to string) (domain.TemplateMapping, error) {
	var payload []byte
	err := r.DB.queryer(ctx).QueryRow(ctx, `SELECT rules FROM template_mappings WHERE from_version_id=$1 AND to_version_id=$2`, from, to).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TemplateMapping{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.TemplateMapping{}, err
	}
	var rules []domain.MappingRule
	if err := json.Unmarshal(payload, &rules); err != nil {
		return domain.TemplateMapping{}, err
	}
	return domain.TemplateMapping{FromTemplateVersionID: from, ToTemplateVersionID: to, Rules: rules}, nil
}
func (r MappingRepository) Put(ctx context.Context, item domain.TemplateMapping) error {
	payload, _ := json.Marshal(item.Rules)
	_, err := r.DB.queryer(ctx).Exec(ctx, `INSERT INTO template_mappings(from_version_id,to_version_id,rules) VALUES($1,$2,$3) ON CONFLICT(from_version_id,to_version_id) DO UPDATE SET rules=EXCLUDED.rules`, item.FromTemplateVersionID, item.ToTemplateVersionID, payload)
	return err
}

type ProfileRepository struct{ DB *Database }

func (r ProfileRepository) Get(ctx context.Context, owner string) (domain.Profile, error) {
	var payload []byte
	var version int64
	err := r.DB.queryer(ctx).QueryRow(ctx, `SELECT payload,version FROM profiles WHERE owner_id=$1`, owner).Scan(&payload, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Profile{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Profile{}, err
	}
	var item domain.Profile
	if err := json.Unmarshal(payload, &item); err != nil {
		return domain.Profile{}, err
	}
	item.Version = version
	return item, nil
}
func (r ProfileRepository) Save(ctx context.Context, item domain.Profile, expected int64) (domain.Profile, error) {
	payload, _ := json.Marshal(item)
	var version int64
	err := r.DB.queryer(ctx).QueryRow(ctx, `INSERT INTO profiles(owner_id,payload,version) VALUES($1,$2,1) ON CONFLICT(owner_id) DO UPDATE SET payload=EXCLUDED.payload,version=profiles.version+1 WHERE profiles.version=$3 RETURNING version`, item.OwnerID, payload, expected).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Profile{}, domain.ErrConflict
	}
	if err != nil {
		return domain.Profile{}, err
	}
	item.Version = version
	return item, nil
}

type PrivacyRepository struct{ DB *Database }

func (r PrivacyRepository) Get(ctx context.Context, owner string) (domain.PrivacyPolicy, int64, error) {
	var payload []byte
	var version int64
	err := r.DB.queryer(ctx).QueryRow(ctx, `SELECT payload,version FROM privacy_policies WHERE owner_id=$1`, owner).Scan(&payload, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PrivacyPolicy{}, 0, domain.ErrNotFound
	}
	if err != nil {
		return domain.PrivacyPolicy{}, 0, err
	}
	var item domain.PrivacyPolicy
	if err := json.Unmarshal(payload, &item); err != nil {
		return domain.PrivacyPolicy{}, 0, err
	}
	return item, version, nil
}
func (r PrivacyRepository) Save(ctx context.Context, item domain.PrivacyPolicy, expected int64) (int64, error) {
	payload, _ := json.Marshal(item)
	var version int64
	err := r.DB.queryer(ctx).QueryRow(ctx, `INSERT INTO privacy_policies(owner_id,payload,version) VALUES($1,$2,1) ON CONFLICT(owner_id) DO UPDATE SET payload=EXCLUDED.payload,version=privacy_policies.version+1 WHERE privacy_policies.version=$3 RETURNING version`, item.OwnerID, payload, expected).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrConflict
	}
	return version, err
}

type ExportRepository struct{ DB *Database }

func (r ExportRepository) FindByIdempotency(ctx context.Context, owner, key string) (domain.ExportRequest, domain.ExportResult, error) {
	var requestPayload, resultPayload []byte
	err := r.DB.queryer(ctx).QueryRow(ctx, `SELECT request_payload,result_payload FROM exports WHERE owner_id=$1 AND idempotency_key=$2`, owner, key).Scan(&requestPayload, &resultPayload)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ExportRequest{}, domain.ExportResult{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ExportRequest{}, domain.ExportResult{}, err
	}
	return decodeExport(requestPayload, resultPayload)
}
func (r ExportRepository) Create(ctx context.Context, request domain.ExportRequest, result domain.ExportResult) error {
	requestPayload, _ := json.Marshal(request)
	resultPayload, _ := json.Marshal(result)
	_, err := r.DB.queryer(ctx).Exec(ctx, `INSERT INTO exports(id,owner_id,idempotency_key,request_payload,result_payload,created_at) VALUES($1,$2,$3,$4,$5,$6)`, result.ID, request.OwnerID, request.IdempotencyKey, requestPayload, resultPayload, result.CreatedAt)
	return err
}
func (r ExportRepository) Get(ctx context.Context, id string) (domain.ExportRequest, domain.ExportResult, error) {
	var requestPayload, resultPayload []byte
	err := r.DB.queryer(ctx).QueryRow(ctx, `SELECT request_payload,result_payload FROM exports WHERE id=$1`, id).Scan(&requestPayload, &resultPayload)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ExportRequest{}, domain.ExportResult{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ExportRequest{}, domain.ExportResult{}, err
	}
	return decodeExport(requestPayload, resultPayload)
}
func decodeExport(requestPayload, resultPayload []byte) (domain.ExportRequest, domain.ExportResult, error) {
	var request domain.ExportRequest
	var result domain.ExportResult
	if err := json.Unmarshal(requestPayload, &request); err != nil {
		return request, result, err
	}
	if err := json.Unmarshal(resultPayload, &result); err != nil {
		return request, result, err
	}
	return request, result, nil
}

type FeedbackRepository struct{ DB *Database }

func (r FeedbackRepository) Create(ctx context.Context, item domain.TemplateFeedback) error {
	_, err := r.DB.queryer(ctx).Exec(ctx, `INSERT INTO template_feedback(id,template_version_id,reporter_id,kind,message,status,resolution,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, item.ID, item.TemplateVersionID, item.ReporterID, item.Kind, item.Message, item.Status, item.Resolution, item.CreatedAt, item.UpdatedAt)
	return err
}
func (r FeedbackRepository) Get(ctx context.Context, id string) (domain.TemplateFeedback, error) {
	var item domain.TemplateFeedback
	err := r.DB.queryer(ctx).QueryRow(ctx, `SELECT id,template_version_id,reporter_id,kind,message,status,resolution,created_at,updated_at FROM template_feedback WHERE id=$1`, id).Scan(&item.ID, &item.TemplateVersionID, &item.ReporterID, &item.Kind, &item.Message, &item.Status, &item.Resolution, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TemplateFeedback{}, domain.ErrNotFound
	}
	return item, err
}
func (r FeedbackRepository) List(ctx context.Context, filter application.FeedbackListFilter) ([]domain.TemplateFeedback, int, error) {
	filter.PageRequest = filter.PageRequest.Normalize(map[string]bool{"updated_at": true, "created_at": true, "status": true})
	orderBy := map[string]string{"updated_at": "updated_at", "created_at": "created_at", "status": "status"}[filter.Sort]
	direction := "DESC"
	if filter.Order == "asc" {
		direction = "ASC"
	}
	query := fmt.Sprintf(`SELECT id,template_version_id,reporter_id,kind,message,status,resolution,created_at,updated_at,count(*) OVER() FROM template_feedback WHERE ($1='' OR status=$1) AND ($2='' OR template_version_id=$2) ORDER BY %s %s LIMIT $3 OFFSET $4`, orderBy, direction)
	rows, err := r.DB.queryer(ctx).Query(ctx, query, filter.Status, filter.TemplateVersionID, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.TemplateFeedback, 0)
	total := 0
	for rows.Next() {
		var item domain.TemplateFeedback
		if err := rows.Scan(&item.ID, &item.TemplateVersionID, &item.ReporterID, &item.Kind, &item.Message, &item.Status, &item.Resolution, &item.CreatedAt, &item.UpdatedAt, &total); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}
func (r FeedbackRepository) Update(ctx context.Context, item domain.TemplateFeedback) error {
	tag, err := r.DB.queryer(ctx).Exec(ctx, `UPDATE template_feedback SET status=$2,resolution=$3,updated_at=$4 WHERE id=$1`, item.ID, item.Status, item.Resolution, item.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return domain.ErrNotFound
	}
	return nil
}

type AttachmentRepository struct{ DB *Database }

func (r AttachmentRepository) Create(ctx context.Context, item domain.Attachment) error {
	_, err := r.DB.queryer(ctx).Exec(ctx, `INSERT INTO attachments(id,owner_id,name,media_type,size,sha256,path,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, item.ID, item.OwnerID, item.Name, item.MediaType, item.Size, item.SHA256, item.Path, item.CreatedAt)
	return err
}
func (r AttachmentRepository) Get(ctx context.Context, id string) (domain.Attachment, error) {
	var item domain.Attachment
	err := r.DB.queryer(ctx).QueryRow(ctx, `SELECT id,owner_id,name,media_type,size,sha256,path,created_at FROM attachments WHERE id=$1`, id).Scan(&item.ID, &item.OwnerID, &item.Name, &item.MediaType, &item.Size, &item.SHA256, &item.Path, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Attachment{}, domain.ErrNotFound
	}
	return item, err
}

type AuditRepository struct{ DB *Database }

func (r AuditRepository) Append(ctx context.Context, item domain.AuditEvent) error {
	metadata, _ := json.Marshal(item.Metadata)
	_, err := r.DB.queryer(ctx).Exec(ctx, `INSERT INTO audit_events(id,actor_id,action,resource,resource_id,request_id,metadata,draft_version,template_version_id,outcome,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, item.ID, item.ActorID, item.Action, item.Resource, item.ResourceID, item.RequestID, metadata, item.DraftVersion, item.TemplateVersionID, item.Outcome, item.CreatedAt)
	return err
}
func (r AuditRepository) List(ctx context.Context, filter application.AuditListFilter) ([]domain.AuditEvent, int, error) {
	filter.PageRequest = filter.PageRequest.Normalize(map[string]bool{"created_at": true, "updated_at": true})
	rows, err := r.DB.queryer(ctx).Query(ctx, `SELECT id,actor_id,action,resource,resource_id,request_id,metadata,draft_version,template_version_id,outcome,created_at,count(*) OVER() FROM audit_events WHERE ($1='' OR actor_id=$1) AND ($2='' OR resource=$2) AND ($3='' OR resource_id=$3) AND ($4='' OR action=$4) ORDER BY created_at DESC LIMIT $5 OFFSET $6`, filter.ActorID, filter.Resource, filter.ResourceID, filter.Action, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.AuditEvent, 0)
	total := 0
	for rows.Next() {
		var item domain.AuditEvent
		var metadata []byte
		if err := rows.Scan(&item.ID, &item.ActorID, &item.Action, &item.Resource, &item.ResourceID, &item.RequestID, &metadata, &item.DraftVersion, &item.TemplateVersionID, &item.Outcome, &item.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(metadata, &item.Metadata)
		items = append(items, item)
	}
	return items, total, rows.Err()
}
