package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
)

type DraftRepository struct{ DB *Database }

func (r DraftRepository) Create(ctx context.Context, item domain.Draft, key, hash string) (domain.Draft, error) {
	var result domain.Draft
	err := r.DB.WithinTransaction(ctx, func(txCtx context.Context) error {
		var existingHash, resourceID string
		err := r.DB.queryer(txCtx).QueryRow(txCtx, `SELECT request_hash,resource_id FROM idempotency_keys WHERE owner_id=$1 AND key=$2`, item.OwnerID, key).Scan(&existingHash, &resourceID)
		if err == nil {
			if existingHash != hash {
				return domain.ErrIdempotencyReuse
			}
			var getErr error
			result, getErr = r.Get(txCtx, resourceID)
			return getErr
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		values, _ := json.Marshal(item.Values)
		unmapped, _ := json.Marshal(item.Unmapped)
		if _, err := r.DB.queryer(txCtx).Exec(txCtx, `INSERT INTO drafts(id,owner_id,target_id,template_version_id,values,unmapped,version,status,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, item.ID, item.OwnerID, item.TargetID, item.TemplateVersionID, values, unmapped, item.Version, item.Status, item.UpdatedAt); err != nil {
			return err
		}
		if _, err := r.DB.queryer(txCtx).Exec(txCtx, `INSERT INTO idempotency_keys(owner_id,key,request_hash,resource_type,resource_id) VALUES($1,$2,$3,'draft',$4)`, item.OwnerID, key, hash, item.ID); err != nil {
			return err
		}
		result = item
		return nil
	})
	return result, err
}
func (r DraftRepository) Get(ctx context.Context, id string) (domain.Draft, error) {
	row := r.DB.queryer(ctx).QueryRow(ctx, `SELECT id,owner_id,target_id,template_version_id,values,unmapped,version,status,updated_at FROM drafts WHERE id=$1`, id)
	return scanDraft(row)
}
func (r DraftRepository) List(ctx context.Context, filter application.DraftListFilter) ([]domain.Draft, int, error) {
	filter.PageRequest = filter.PageRequest.Normalize(map[string]bool{"updated_at": true, "created_at": true})
	direction := "DESC"
	if filter.Order == "asc" {
		direction = "ASC"
	}
	query := fmt.Sprintf(`SELECT id,owner_id,target_id,template_version_id,values,unmapped,version,status,updated_at,count(*) OVER() FROM drafts WHERE ($1='' OR owner_id=$1) AND ($2='' OR status=$2) AND ($3='' OR target_id=$3) ORDER BY updated_at %s LIMIT $4 OFFSET $5`, direction)
	rows, err := r.DB.queryer(ctx).Query(ctx, query, filter.OwnerID, filter.Status, filter.TargetID, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.Draft, 0)
	total := 0
	for rows.Next() {
		var item domain.Draft
		var values, unmapped []byte
		if err := rows.Scan(&item.ID, &item.OwnerID, &item.TargetID, &item.TemplateVersionID, &values, &unmapped, &item.Version, &item.Status, &item.UpdatedAt, &total); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(values, &item.Values); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(unmapped, &item.Unmapped); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}
func (r DraftRepository) Save(ctx context.Context, item domain.Draft, expected int64, key, hash string) (domain.Draft, error) {
	var result domain.Draft
	err := r.DB.WithinTransaction(ctx, func(txCtx context.Context) error {
		var existingHash, resourceID string
		err := r.DB.queryer(txCtx).QueryRow(txCtx, `SELECT request_hash,resource_id FROM idempotency_keys WHERE owner_id=$1 AND key=$2`, item.OwnerID, key).Scan(&existingHash, &resourceID)
		if err == nil {
			if existingHash != hash {
				return domain.ErrIdempotencyReuse
			}
			var getErr error
			result, getErr = r.Get(txCtx, resourceID)
			return getErr
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		values, _ := json.Marshal(item.Values)
		unmapped, _ := json.Marshal(item.Unmapped)
		tag, err := r.DB.queryer(txCtx).Exec(txCtx, `UPDATE drafts SET template_version_id=$2,values=$3,unmapped=$4,version=$5,status=$6,updated_at=$7 WHERE id=$1 AND version=$8`, item.ID, item.TemplateVersionID, values, unmapped, item.Version, item.Status, item.UpdatedAt, expected)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return domain.ErrConflict
		}
		if _, err := r.DB.queryer(txCtx).Exec(txCtx, `INSERT INTO idempotency_keys(owner_id,key,request_hash,resource_type,resource_id) VALUES($1,$2,$3,'draft',$4)`, item.OwnerID, key, hash, item.ID); err != nil {
			return err
		}
		result = item
		return nil
	})
	return result, err
}
func (r DraftRepository) Archive(ctx context.Context, id string, expected int64, now time.Time) error {
	tag, err := r.DB.queryer(ctx).Exec(ctx, `UPDATE drafts SET status='archived',version=version+1,updated_at=$3 WHERE id=$1 AND version=$2`, id, expected, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}
func scanDraft(row pgx.Row) (domain.Draft, error) {
	var item domain.Draft
	var values, unmapped []byte
	err := row.Scan(&item.ID, &item.OwnerID, &item.TargetID, &item.TemplateVersionID, &values, &unmapped, &item.Version, &item.Status, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Draft{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Draft{}, err
	}
	if err := json.Unmarshal(values, &item.Values); err != nil {
		return domain.Draft{}, err
	}
	if err := json.Unmarshal(unmapped, &item.Unmapped); err != nil {
		return domain.Draft{}, err
	}
	return item, nil
}
