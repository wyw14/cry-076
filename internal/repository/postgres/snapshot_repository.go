package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
)

type SnapshotRepository struct{ DB *Database }

func (r SnapshotRepository) Create(ctx context.Context, item domain.DraftSnapshot) error {
	values, _ := json.Marshal(item.Values)
	unmapped, _ := json.Marshal(item.Unmapped)
	_, err := r.DB.queryer(ctx).Exec(ctx, `INSERT INTO draft_snapshots(id,draft_id,version,template_version_id,values,unmapped,reason,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, item.ID, item.DraftID, item.Version, item.TemplateVersionID, values, unmapped, item.Reason, item.CreatedAt)
	return err
}
func (r SnapshotRepository) Get(ctx context.Context, id string) (domain.DraftSnapshot, error) {
	var item domain.DraftSnapshot
	var values, unmapped []byte
	err := r.DB.queryer(ctx).QueryRow(ctx, `SELECT id,draft_id,version,template_version_id,values,unmapped,reason,created_at FROM draft_snapshots WHERE id=$1`, id).Scan(&item.ID, &item.DraftID, &item.Version, &item.TemplateVersionID, &values, &unmapped, &item.Reason, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DraftSnapshot{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DraftSnapshot{}, err
	}
	if err := json.Unmarshal(values, &item.Values); err != nil {
		return domain.DraftSnapshot{}, err
	}
	if err := json.Unmarshal(unmapped, &item.Unmapped); err != nil {
		return domain.DraftSnapshot{}, err
	}
	return item, nil
}
func (r SnapshotRepository) ListByDraft(ctx context.Context, draftID string, page application.PageRequest) ([]domain.DraftSnapshot, int, error) {
	page = page.Normalize(map[string]bool{"created_at": true, "updated_at": true})
	rows, err := r.DB.queryer(ctx).Query(ctx, `SELECT id,draft_id,version,template_version_id,values,unmapped,reason,created_at,count(*) OVER() FROM draft_snapshots WHERE draft_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, draftID, page.PageSize, (page.Page-1)*page.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.DraftSnapshot, 0)
	total := 0
	for rows.Next() {
		var item domain.DraftSnapshot
		var values, unmapped []byte
		if err := rows.Scan(&item.ID, &item.DraftID, &item.Version, &item.TemplateVersionID, &values, &unmapped, &item.Reason, &item.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(values, &item.Values)
		_ = json.Unmarshal(unmapped, &item.Unmapped)
		items = append(items, item)
	}
	return items, total, rows.Err()
}
