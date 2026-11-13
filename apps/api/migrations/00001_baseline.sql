-- +goose Up
-- baseline: the production schema as GORM AutoMigrate and schemas.MigrateWithIssuer built it, captured 2026-11-13 with pg_dump --schema-only --no-owner --no-privileges

CREATE TABLE api_keys (
    id bigint NOT NULL,
    user_id bigint,
    site_id bigint,
    name text,
    prefix text,
    key_hint text,
    key_hash text,
    scopes text DEFAULT 'read'::text,
    is_active boolean DEFAULT true,
    last_used_at timestamp with time zone,
    expires_at timestamp with time zone,
    created_at timestamp with time zone,
    workspace_id bigint
);

CREATE SEQUENCE api_keys_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE api_keys_id_seq OWNED BY api_keys.id;

CREATE TABLE custom_events (
    id bigint NOT NULL,
    site_id bigint,
    visitor_id text,
    path text,
    name text,
    props text,
    created_at timestamp with time zone
);

CREATE SEQUENCE custom_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE custom_events_id_seq OWNED BY custom_events.id;

CREATE TABLE directory_deletion_state (
    id smallint NOT NULL,
    state bytea NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT directory_deletion_state_id_check CHECK ((id = 1))
);

CREATE TABLE goals (
    id bigint NOT NULL,
    site_id bigint,
    owner_id bigint,
    name text,
    goal_type text,
    event_name text,
    page_path text,
    match_type text DEFAULT 'exact'::text,
    created_at timestamp with time zone
);

CREATE SEQUENCE goals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE goals_id_seq OWNED BY goals.id;

CREATE TABLE mcp_actions (
    id text NOT NULL,
    user_id text NOT NULL,
    kind text NOT NULL,
    target text NOT NULL,
    digest text NOT NULL,
    summary text DEFAULT ''::text NOT NULL,
    client_id text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    resolved_at timestamp with time zone,
    outcome text DEFAULT ''::text NOT NULL
);

CREATE TABLE pageviews (
    id bigint NOT NULL,
    site_id bigint,
    path text,
    referrer text,
    user_agent text,
    language text,
    country text,
    visitor_id text,
    created_at timestamp with time zone,
    browser text,
    os text,
    device text,
    screen_width bigint,
    utm_source text,
    utm_medium text,
    utm_campaign text,
    utm_term text,
    utm_content text,
    perf_dns bigint,
    perf_tcp bigint,
    perf_ttfb bigint,
    perf_dom_load bigint,
    perf_page_load bigint
);

CREATE SEQUENCE pageviews_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE pageviews_id_seq OWNED BY pageviews.id;

CREATE TABLE porte_identities (
    user_id bigint NOT NULL,
    provider text NOT NULL,
    subject text NOT NULL,
    password_hash text DEFAULT ''::text NOT NULL,
    access_token text DEFAULT ''::text NOT NULL,
    refresh_token text DEFAULT ''::text NOT NULL,
    token_expiry timestamp with time zone,
    roles jsonb,
    roles_synced_at timestamp with time zone,
    synced_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE porte_login_codes (
    code_hash text NOT NULL,
    user_id bigint NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    provider text DEFAULT ''::text NOT NULL,
    subject text DEFAULT ''::text NOT NULL
);

CREATE TABLE porte_magic_links (
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    ip text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone
);

CREATE TABLE porte_sessions (
    id bigint NOT NULL,
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    label text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone
);

CREATE SEQUENCE porte_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE porte_sessions_id_seq OWNED BY porte_sessions.id;

CREATE TABLE sites (
    id bigint NOT NULL,
    name text,
    domain text,
    owner_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    share_token text,
    workspace_id bigint
);

CREATE SEQUENCE sites_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE sites_id_seq OWNED BY sites.id;

CREATE TABLE users (
    id bigint NOT NULL,
    email text,
    name text,
    password_hash text,
    created_at timestamp with time zone,
    avatar_url text,
    avatar_source text,
    oidc_picture_url text,
    oidc_access_token text,
    oidc_refresh_token text,
    oidc_token_expiry timestamp with time zone,
    profile_synced_at timestamp with time zone,
    oidc_subject text,
    locale text DEFAULT 'en'::text NOT NULL
);

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE users_id_seq OWNED BY users.id;

CREATE TABLE visitor_sessions (
    id bigint NOT NULL,
    site_id bigint,
    visitor_id text,
    started_at timestamp with time zone,
    ended_at timestamp with time zone,
    entry_path text,
    exit_path text,
    pageview_count bigint DEFAULT 1,
    duration bigint DEFAULT 0,
    is_bounce boolean DEFAULT true,
    created_at timestamp with time zone
);

CREATE SEQUENCE visitor_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitor_sessions_id_seq OWNED BY visitor_sessions.id;

CREATE TABLE webhooks (
    id bigint NOT NULL,
    site_id bigint,
    url text,
    secret text,
    period text,
    enabled boolean DEFAULT true,
    last_sent_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    owner_id bigint,
    interval_hours bigint DEFAULT 24,
    workspace_id bigint
);

CREATE SEQUENCE webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE webhooks_id_seq OWNED BY webhooks.id;

CREATE TABLE workspace_members (
    id bigint NOT NULL,
    workspace_id bigint,
    user_id bigint,
    role text,
    created_at timestamp with time zone
);

CREATE SEQUENCE workspace_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE workspace_members_id_seq OWNED BY workspace_members.id;

CREATE TABLE workspaces (
    id bigint NOT NULL,
    name text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    registre_space_id text
);

CREATE SEQUENCE workspaces_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE workspaces_id_seq OWNED BY workspaces.id;

ALTER TABLE ONLY api_keys ALTER COLUMN id SET DEFAULT nextval('api_keys_id_seq'::regclass);

ALTER TABLE ONLY custom_events ALTER COLUMN id SET DEFAULT nextval('custom_events_id_seq'::regclass);

ALTER TABLE ONLY goals ALTER COLUMN id SET DEFAULT nextval('goals_id_seq'::regclass);

ALTER TABLE ONLY pageviews ALTER COLUMN id SET DEFAULT nextval('pageviews_id_seq'::regclass);

ALTER TABLE ONLY porte_sessions ALTER COLUMN id SET DEFAULT nextval('porte_sessions_id_seq'::regclass);

ALTER TABLE ONLY sites ALTER COLUMN id SET DEFAULT nextval('sites_id_seq'::regclass);

ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);

ALTER TABLE ONLY visitor_sessions ALTER COLUMN id SET DEFAULT nextval('visitor_sessions_id_seq'::regclass);

ALTER TABLE ONLY webhooks ALTER COLUMN id SET DEFAULT nextval('webhooks_id_seq'::regclass);

ALTER TABLE ONLY workspace_members ALTER COLUMN id SET DEFAULT nextval('workspace_members_id_seq'::regclass);

ALTER TABLE ONLY workspaces ALTER COLUMN id SET DEFAULT nextval('workspaces_id_seq'::regclass);

ALTER TABLE ONLY api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);

ALTER TABLE ONLY custom_events
    ADD CONSTRAINT custom_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY directory_deletion_state
    ADD CONSTRAINT directory_deletion_state_pkey PRIMARY KEY (id);

ALTER TABLE ONLY goals
    ADD CONSTRAINT goals_pkey PRIMARY KEY (id);

ALTER TABLE ONLY mcp_actions
    ADD CONSTRAINT mcp_actions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY pageviews
    ADD CONSTRAINT pageviews_pkey PRIMARY KEY (id);

ALTER TABLE ONLY porte_identities
    ADD CONSTRAINT porte_identities_pkey PRIMARY KEY (provider, subject);

ALTER TABLE ONLY porte_login_codes
    ADD CONSTRAINT porte_login_codes_pkey PRIMARY KEY (code_hash);

ALTER TABLE ONLY porte_magic_links
    ADD CONSTRAINT porte_magic_links_pkey PRIMARY KEY (token_hash);

ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_token_hash_key UNIQUE (token_hash);

ALTER TABLE ONLY sites
    ADD CONSTRAINT sites_pkey PRIMARY KEY (id);

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY visitor_sessions
    ADD CONSTRAINT visitor_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY workspace_members
    ADD CONSTRAINT workspace_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY workspaces
    ADD CONSTRAINT workspaces_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX idx_api_keys_key_hash ON api_keys USING btree (key_hash);

CREATE INDEX idx_api_keys_site_id ON api_keys USING btree (site_id);

CREATE INDEX idx_api_keys_user_id ON api_keys USING btree (user_id);

CREATE INDEX idx_api_keys_workspace_id ON api_keys USING btree (workspace_id);

CREATE INDEX idx_custom_events_name ON custom_events USING btree (name);

CREATE INDEX idx_event_site_created ON custom_events USING btree (site_id, created_at);

CREATE INDEX idx_goals_owner_id ON goals USING btree (owner_id);

CREATE INDEX idx_goals_site_id ON goals USING btree (site_id);

CREATE INDEX idx_pageviews_created_at ON pageviews USING btree (created_at);

CREATE INDEX idx_pageviews_path ON pageviews USING btree (path);

CREATE INDEX idx_pageviews_site_id ON pageviews USING btree (site_id);

CREATE INDEX idx_pageviews_visitor_id ON pageviews USING btree (visitor_id);

CREATE INDEX idx_site_created ON pageviews USING btree (site_id, created_at);

CREATE UNIQUE INDEX idx_sites_domain ON sites USING btree (domain);

CREATE INDEX idx_sites_owner_id ON sites USING btree (owner_id);

CREATE UNIQUE INDEX idx_sites_share_token ON sites USING btree (share_token);

CREATE INDEX idx_sites_workspace_id ON sites USING btree (workspace_id);

CREATE UNIQUE INDEX idx_users_email ON users USING btree (email);

CREATE UNIQUE INDEX idx_users_o_id_c_subject ON users USING btree (oidc_subject);

CREATE INDEX idx_visitor_sessions_site_id ON visitor_sessions USING btree (site_id);

CREATE INDEX idx_webhooks_owner_id ON webhooks USING btree (owner_id);

CREATE INDEX idx_webhooks_site_id ON webhooks USING btree (site_id);

CREATE INDEX idx_webhooks_workspace_id ON webhooks USING btree (workspace_id);

CREATE INDEX idx_workspace_members_user_id ON workspace_members USING btree (user_id);

CREATE INDEX idx_workspace_members_workspace_id ON workspace_members USING btree (workspace_id);

CREATE UNIQUE INDEX idx_workspaces_registre_space_id ON workspaces USING btree (registre_space_id);

CREATE UNIQUE INDEX idx_ws_user ON workspace_members USING btree (workspace_id, user_id);

CREATE INDEX mcp_actions_open_idx ON mcp_actions USING btree (user_id, created_at DESC) WHERE (resolved_at IS NULL);

CREATE INDEX porte_identities_user_idx ON porte_identities USING btree (user_id);

CREATE INDEX porte_magic_links_user_idx ON porte_magic_links USING btree (user_id, created_at);

CREATE INDEX porte_sessions_expiry_idx ON porte_sessions USING btree (expires_at) WHERE (expires_at IS NOT NULL);

CREATE INDEX porte_sessions_user_idx ON porte_sessions USING btree (user_id);

ALTER TABLE ONLY workspace_members
    ADD CONSTRAINT fk_workspace_members_user FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE ONLY workspace_members
    ADD CONSTRAINT fk_workspace_members_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id);

ALTER TABLE ONLY porte_identities
    ADD CONSTRAINT porte_identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY porte_login_codes
    ADD CONSTRAINT porte_login_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY porte_magic_links
    ADD CONSTRAINT porte_magic_links_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
SELECT 'baseline migration is not reversible' WHERE false;
