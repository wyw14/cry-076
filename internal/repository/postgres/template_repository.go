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

type TemplateRepository struct{ DB *Database }

func (r TemplateRepository) CreateVersion(ctx context.Context, item domain.TemplateVersion) error {
	sections, _ := json.Marshal(item.Sections)
	scenarios, _ := json.Marshal(item.Scenarios)
	style, _ := json.Marshal(item.Style)
	_, err := r.DB.queryer(ctx).Exec(ctx, `INSERT INTO template_versions(id,template_id,version,name,category,scenarios,status,sections,style,published_at,deprecated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, item.ID, item.TemplateID, item.Version, item.Name, item.Category, scenarios, item.Status, sections, style, item.PublishedAt, item.DeprecatedAt)
	return err
}
func (r TemplateRepository) GetVersion(ctx context.Context, id string) (domain.TemplateVersion, error) {
	row := r.DB.queryer(ctx).QueryRow(ctx, `SELECT id,template_id,version,name,category,scenarios,status,sections,style,published_at,deprecated_at FROM template_versions WHERE id=$1`, id)
	return scanTemplate(row)
}
func (r TemplateRepository) ListVersions(ctx context.Context, filter application.TemplateListFilter) ([]domain.TemplateVersion, int, error) {
	filter.PageRequest = filter.PageRequest.Normalize(map[string]bool{"updated_at": true, "name": true, "version": true})
	orderBy := map[string]string{"updated_at": "created_at", "name": "name", "version": "version"}[filter.Sort]
	direction := "DESC"
	if filter.Order == "asc" {
		direction = "ASC"
	}
	query := fmt.Sprintf(`SELECT id,template_id,version,name,category,scenarios,status,sections,style,published_at,deprecated_at,count(*) OVER() FROM template_versions WHERE ($1='' OR status=$1) AND ($2='' OR category=$2) AND ($3='' OR scenarios @> to_jsonb(ARRAY[$3]::text[])) ORDER BY %s %s LIMIT $4 OFFSET $5`, orderBy, direction)
	rows, err := r.DB.queryer(ctx).Query(ctx, query, filter.Status, filter.Category, filter.Scenario, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.TemplateVersion, 0)
	total := 0
	for rows.Next() {
		var item domain.TemplateVersion
		var scenarios, sections, style []byte
		if err := rows.Scan(&item.ID, &item.TemplateID, &item.Version, &item.Name, &item.Category, &scenarios, &item.Status, &sections, &style, &item.PublishedAt, &item.DeprecatedAt, &total); err != nil {
			return nil, 0, err
		}
		if err := decodeTemplateJSON(&item, scenarios, sections, style); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}
func (r TemplateRepository) UpdateVersion(ctx context.Context, item domain.TemplateVersion, expected int64) error {
	sections, _ := json.Marshal(item.Sections)
	scenarios, _ := json.Marshal(item.Scenarios)
	style, _ := json.Marshal(item.Style)
	tag, err := r.DB.queryer(ctx).Exec(ctx, `UPDATE template_versions SET version=$2,name=$3,category=$4,scenarios=$5,status=$6,sections=$7,style=$8,published_at=$9,deprecated_at=$10 WHERE id=$1 AND version=$11`, item.ID, item.Version, item.Name, item.Category, scenarios, item.Status, sections, style, item.PublishedAt, item.DeprecatedAt, expected)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}
func scanTemplate(row pgx.Row) (domain.TemplateVersion, error) {
	var item domain.TemplateVersion
	var scenarios, sections, style []byte
	err := row.Scan(&item.ID, &item.TemplateID, &item.Version, &item.Name, &item.Category, &scenarios, &item.Status, &sections, &style, &item.PublishedAt, &item.DeprecatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TemplateVersion{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.TemplateVersion{}, err
	}
	if err := decodeTemplateJSON(&item, scenarios, sections, style); err != nil {
		return domain.TemplateVersion{}, err
	}
	return item, nil
}
func decodeTemplateJSON(item *domain.TemplateVersion, scenarios, sections, style []byte) error {
	if err := json.Unmarshal(scenarios, &item.Scenarios); err != nil {
		return err
	}
	if err := json.Unmarshal(sections, &item.Sections); err != nil {
		return err
	}
	return json.Unmarshal(style, &item.Style)
}
