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
	pipeline := newRenderPipeline(context.WithoutCancel(ctx), layout, draft, privacy)
	if err := pipeline.collect(); err != nil {
		return nil, nil, err
	}
	content, err := pipeline.encode(format)
	if err != nil {
		return nil, nil, err
	}
	return content, pipeline.fieldKeys(), nil
}

type renderPipeline struct {
	ctx     context.Context
	layout  domain.TemplateVersion
	draft   domain.Draft
	privacy domain.PrivacyPolicy
	fields  []renderField
}

func newRenderPipeline(ctx context.Context, layout domain.TemplateVersion, draft domain.Draft, privacy domain.PrivacyPolicy) *renderPipeline {
	return &renderPipeline{
		ctx: ctx, layout: layout, draft: draft, privacy: privacy,
		fields: make([]renderField, 0, len(layout.Fields())),
	}
}

func (p *renderPipeline) collect() error {
	for _, spec := range p.layout.Fields() {
		if err := p.ctx.Err(); err != nil {
			return err
		}
		field, included := p.materialize(spec)
		if !included {
			continue
		}
		p.fields = append(p.fields, field)
	}
	return nil
}

func (p *renderPipeline) materialize(spec domain.FieldSpec) (renderField, bool) {
	if !p.privacy.Includes(spec.Key) {
		return renderField{}, false
	}
	value, exists := p.draft.Values[spec.Key]
	if !exists {
		return renderField{}, false
	}
	if !spec.Visibility.Visible(domain.ScenarioSocial, p.draft.Values) {
		return renderField{}, false
	}
	return renderField{Key: spec.Key, Label: spec.Label, Value: value}, true
}

func (p *renderPipeline) fieldKeys() []string {
	keys := make([]string, 0, len(p.fields))
	for _, field := range p.fields {
		keys = append(keys, field.Key)
	}
	sort.Strings(keys)
	return keys
}

func (p *renderPipeline) encode(format domain.ExportFormat) ([]byte, error) {
	switch format {
	case domain.ExportJSON:
		return json.MarshalIndent(map[string]any{
			"template": p.layout.Name,
			"fields":   p.fields,
		}, "", "  ")
	case domain.ExportHTML:
		return p.encodeHTML()
	default:
		return nil, fmt.Errorf("%w: unsupported export format", domain.ErrValidation)
	}
}

func (p *renderPipeline) encodeHTML() ([]byte, error) {
	page, err := template.New("resume").Parse(`<!doctype html><html><head><meta charset="utf-8"><style>body{font-family:{{.Style.FontFamily}}}h1{color:{{.Style.AccentColor}}}</style></head><body><h1>{{.Name}}</h1>{{range .Fields}}<section><h2>{{.Label}}</h2><div>{{.Value}}</div></section>{{end}}</body></html>`)
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	if err := page.Execute(&output, renderDocument{Name: p.layout.Name, Style: p.layout.Style, Fields: p.fields}); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
