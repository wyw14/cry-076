INSERT INTO template_versions (
    id,
    template_id,
    version,
    name,
    category,
    scenarios,
    status,
    sections,
    style,
    published_at
) VALUES (
    'tv_demo_modern_1',
    'template_modern',
    1,
    '清晰双栏',
    '通用',
    '["campus", "social", "technical"]'::jsonb,
    'published',
    '[
      {
        "key": "identity",
        "title": "基本资料",
        "order": 10,
        "repeatable": false,
        "fields": [
          {"key": "full_name", "label": "姓名", "kind": "text", "required": true, "private": false, "visibility": {}},
          {"key": "email", "label": "邮箱", "kind": "text", "required": true, "private": true, "visibility": {}},
          {"key": "phone", "label": "电话", "kind": "text", "required": false, "private": true, "visibility": {}}
        ]
      },
      {
        "key": "experience",
        "title": "工作经历",
        "order": 20,
        "repeatable": true,
        "fields": [
          {"key": "experiences", "label": "经历", "kind": "collection", "required": true, "private": false, "visibility": {}}
        ]
      }
    ]'::jsonb,
    '{
      "page_size": "A4",
      "font_family": "system-ui",
      "accent_color": "#0f766e",
      "density": "comfortable"
    }'::jsonb,
    now()
)
ON CONFLICT (id) DO NOTHING;
