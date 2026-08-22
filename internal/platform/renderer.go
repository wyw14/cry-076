package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"

	"github.com/wyw14/cry-076/internal/domain"
)

type OfflineRenderer struct{}
type renderField struct {
	Key, Label string
	Value      any
}
type renderDocument struct {
	Name   string
	Style  domain.PreviewStyle
	Fields []renderField
}

func (OfflineRenderer) Render(ctx context.Context, layout domain.TemplateVersion, draft domain.Draft, privacy domain.PrivacyPolicy, format domain.ExportFormat) ([]byte, []string, error) {
	fields := make([]renderField, 0)
	keys := make([]string, 0)
	for _, spec := range layout.Fields() {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
		}
		if !privacy.Includes(spec.Key) {
			continue
		}
		value, exists := draft.Values[spec.Key]
		if !exists || !spec.Visibility.Visible(domain.ScenarioSocial, draft.Values) {
			continue
		}
		fields = append(fields, renderField{Key: spec.Key, Label: spec.Label, Value: value})
		keys = append(keys, spec.Key)
	}
	sort.Strings(keys)
	switch format {
	case domain.ExportJSON:
		data, err := json.MarshalIndent(map[string]any{"template": layout.Name, "fields": fields}, "", "  ")
		return data, keys, err
	case domain.ExportHTML:
		page, err := template.New("resume").Parse(`<!doctype html><html><head><meta charset="utf-8"><style>body{font-family:{{.Style.FontFamily}}}h1{color:{{.Style.AccentColor}}}</style></head><body><h1>{{.Name}}</h1>{{range .Fields}}<section><h2>{{.Label}}</h2><div>{{.Value}}</div></section>{{end}}</body></html>`)
		if err != nil {
			return nil, nil, err
		}
		var output bytes.Buffer
		if err := page.Execute(&output, renderDocument{Name: layout.Name, Style: layout.Style, Fields: fields}); err != nil {
			return nil, nil, err
		}
		return output.Bytes(), keys, nil
	default:
		return nil, nil, fmt.Errorf("%w: unsupported export format", domain.ErrValidation)
	}
}
