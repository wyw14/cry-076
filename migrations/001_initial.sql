CREATE TABLE IF NOT EXISTS template_versions (
    id text PRIMARY KEY,
    template_id text NOT NULL,
    version bigint NOT NULL CHECK (version > 0),
    name text NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 120),
    category text NOT NULL CHECK (length(trim(category)) > 0),
    scenarios jsonb NOT NULL CHECK (jsonb_typeof(scenarios) = 'array'),
    status text NOT NULL CHECK (status IN ('draft', 'review', 'published', 'deprecated')),
    sections jsonb NOT NULL CHECK (jsonb_typeof(sections) = 'array'),
    style jsonb NOT NULL CHECK (jsonb_typeof(style) = 'object'),
    published_at timestamptz,
    deprecated_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (template_id, version),
    CHECK (status <> 'published' OR published_at IS NOT NULL),
    CHECK (status <> 'deprecated' OR deprecated_at IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS template_versions_library_idx
    ON template_versions (status, category, version DESC);

CREATE TABLE IF NOT EXISTS template_mappings (
    from_version_id text NOT NULL REFERENCES template_versions(id),
    to_version_id text NOT NULL REFERENCES template_versions(id),
    rules jsonb NOT NULL CHECK (jsonb_typeof(rules) = 'object'),
    PRIMARY KEY (from_version_id, to_version_id),
    CHECK (from_version_id <> to_version_id)
);

CREATE TABLE IF NOT EXISTS drafts (
    id text PRIMARY KEY,
    owner_id text NOT NULL CHECK (length(trim(owner_id)) > 0),
    target_id text NOT NULL CHECK (length(trim(target_id)) > 0),
    template_version_id text NOT NULL REFERENCES template_versions(id),
    values jsonb NOT NULL CHECK (jsonb_typeof(values) = 'object'),
    unmapped jsonb NOT NULL CHECK (jsonb_typeof(unmapped) = 'object'),
    version bigint NOT NULL CHECK (version > 0),
    status text NOT NULL CHECK (status IN ('active', 'archived')),
    updated_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS drafts_owner_workspace_idx
    ON drafts (owner_id, status, updated_at DESC);

CREATE TABLE IF NOT EXISTS draft_snapshots (
    id text PRIMARY KEY,
    draft_id text NOT NULL REFERENCES drafts(id) ON DELETE CASCADE,
    version bigint NOT NULL CHECK (version > 0),
    template_version_id text NOT NULL REFERENCES template_versions(id),
    values jsonb NOT NULL CHECK (jsonb_typeof(values) = 'object'),
    unmapped jsonb NOT NULL CHECK (jsonb_typeof(unmapped) = 'object'),
    reason text NOT NULL CHECK (length(trim(reason)) > 0),
    created_at timestamptz NOT NULL,
    UNIQUE (draft_id, version, reason)
);

CREATE INDEX IF NOT EXISTS draft_snapshots_timeline_idx
    ON draft_snapshots (draft_id, created_at DESC);

CREATE TABLE IF NOT EXISTS profiles (
    owner_id text PRIMARY KEY,
    payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    version bigint NOT NULL CHECK (version > 0)
);

CREATE TABLE IF NOT EXISTS privacy_policies (
    owner_id text PRIMARY KEY,
    payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    version bigint NOT NULL CHECK (version > 0)
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    owner_id text NOT NULL,
    key text NOT NULL,
    request_hash text NOT NULL CHECK (length(request_hash) = 64),
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (owner_id, key)
);

CREATE TABLE IF NOT EXISTS exports (
    id text PRIMARY KEY,
    owner_id text NOT NULL,
    idempotency_key text NOT NULL,
    request_payload jsonb NOT NULL CHECK (jsonb_typeof(request_payload) = 'object'),
    result_payload jsonb NOT NULL CHECK (jsonb_typeof(result_payload) = 'object'),
    created_at timestamptz NOT NULL,
    UNIQUE (owner_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS exports_owner_history_idx
    ON exports (owner_id, created_at DESC);

CREATE TABLE IF NOT EXISTS template_feedback (
    id text PRIMARY KEY,
    template_version_id text NOT NULL REFERENCES template_versions(id),
    reporter_id text NOT NULL,
    kind text NOT NULL CHECK (kind IN ('blocking', 'content', 'style', 'accessibility')),
    message text NOT NULL CHECK (length(trim(message)) >= 4),
    status text NOT NULL CHECK (status IN ('open', 'triaged', 'resolved', 'dismissed')),
    resolution text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS feedback_release_gate_idx
    ON template_feedback (template_version_id, kind, status);

CREATE TABLE IF NOT EXISTS attachments (
    id text PRIMARY KEY,
    owner_id text NOT NULL,
    name text NOT NULL CHECK (length(trim(name)) > 0),
    media_type text NOT NULL,
    size bigint NOT NULL CHECK (size > 0),
    sha256 text NOT NULL CHECK (length(sha256) = 64),
    path text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_events (
    id text PRIMARY KEY,
    actor_id text NOT NULL,
    action text NOT NULL,
    resource text NOT NULL,
    resource_id text NOT NULL,
    request_id text NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(metadata) = 'object'),
    draft_version bigint NOT NULL DEFAULT 0 CHECK (draft_version >= 0),
    template_version_id text NOT NULL DEFAULT '',
    outcome text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS audit_material_access_idx
    ON audit_events (resource, resource_id, action, created_at DESC);
