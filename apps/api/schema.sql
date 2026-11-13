CREATE TABLE public.api_keys (
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
    created_at timestamp with time zone DEFAULT now(),
    workspace_id bigint
);

CREATE SEQUENCE public.api_keys_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.api_keys_id_seq OWNED BY public.api_keys.id;

CREATE TABLE public.custom_events (
    id bigint NOT NULL,
    site_id bigint,
    visitor_id text,
    path text,
    name text,
    props text,
    created_at timestamp with time zone DEFAULT now()
);

CREATE SEQUENCE public.custom_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.custom_events_id_seq OWNED BY public.custom_events.id;

CREATE TABLE public.directory_deletion_state (
    id smallint NOT NULL,
    state bytea NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT directory_deletion_state_id_check CHECK ((id = 1))
);

CREATE TABLE public.goals (
    id bigint NOT NULL,
    site_id bigint,
    owner_id bigint,
    name text,
    goal_type text,
    event_name text,
    page_path text,
    match_type text DEFAULT 'exact'::text,
    created_at timestamp with time zone DEFAULT now()
);

CREATE SEQUENCE public.goals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.goals_id_seq OWNED BY public.goals.id;

CREATE TABLE public.mcp_actions (
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

CREATE TABLE public.pageviews (
    id bigint NOT NULL,
    site_id bigint,
    path text,
    referrer text,
    user_agent text,
    language text,
    country text,
    visitor_id text,
    created_at timestamp with time zone DEFAULT now(),
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

CREATE SEQUENCE public.pageviews_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.pageviews_id_seq OWNED BY public.pageviews.id;

CREATE TABLE public.porte_identities (
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

CREATE TABLE public.porte_login_codes (
    code_hash text NOT NULL,
    user_id bigint NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    provider text DEFAULT ''::text NOT NULL,
    subject text DEFAULT ''::text NOT NULL
);

CREATE TABLE public.porte_magic_links (
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    ip text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone
);

CREATE TABLE public.porte_sessions (
    id bigint NOT NULL,
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    label text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone
);

CREATE SEQUENCE public.porte_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.porte_sessions_id_seq OWNED BY public.porte_sessions.id;

CREATE TABLE public.sites (
    id bigint NOT NULL,
    name text,
    domain text,
    owner_id bigint,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    share_token text,
    workspace_id bigint
);

CREATE SEQUENCE public.sites_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.sites_id_seq OWNED BY public.sites.id;

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text,
    name text,
    password_hash text DEFAULT ''::text,
    created_at timestamp with time zone DEFAULT now(),
    avatar_url text DEFAULT ''::text,
    avatar_source text DEFAULT ''::text,
    oidc_picture_url text,
    oidc_access_token text DEFAULT ''::text,
    oidc_refresh_token text DEFAULT ''::text,
    oidc_token_expiry timestamp with time zone DEFAULT '0001-01-01 00:00:00+00'::timestamp with time zone,
    profile_synced_at timestamp with time zone DEFAULT '0001-01-01 00:00:00+00'::timestamp with time zone,
    oidc_subject text,
    locale text DEFAULT 'en'::text NOT NULL
);

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

CREATE TABLE public.visitor_sessions (
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
    created_at timestamp with time zone DEFAULT now()
);

CREATE SEQUENCE public.visitor_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.visitor_sessions_id_seq OWNED BY public.visitor_sessions.id;

CREATE TABLE public.webhooks (
    id bigint NOT NULL,
    site_id bigint,
    url text,
    secret text,
    period text,
    enabled boolean DEFAULT true,
    last_sent_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    owner_id bigint,
    interval_hours bigint DEFAULT 24,
    workspace_id bigint
);

CREATE SEQUENCE public.webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.webhooks_id_seq OWNED BY public.webhooks.id;

CREATE TABLE public.workspace_members (
    id bigint NOT NULL,
    workspace_id bigint,
    user_id bigint,
    role text,
    created_at timestamp with time zone DEFAULT now()
);

CREATE SEQUENCE public.workspace_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.workspace_members_id_seq OWNED BY public.workspace_members.id;

CREATE TABLE public.workspaces (
    id bigint NOT NULL,
    name text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    registre_space_id text
);

CREATE SEQUENCE public.workspaces_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.workspaces_id_seq OWNED BY public.workspaces.id;

ALTER TABLE ONLY public.api_keys ALTER COLUMN id SET DEFAULT nextval('public.api_keys_id_seq'::regclass);

ALTER TABLE ONLY public.custom_events ALTER COLUMN id SET DEFAULT nextval('public.custom_events_id_seq'::regclass);

ALTER TABLE ONLY public.goals ALTER COLUMN id SET DEFAULT nextval('public.goals_id_seq'::regclass);

ALTER TABLE ONLY public.pageviews ALTER COLUMN id SET DEFAULT nextval('public.pageviews_id_seq'::regclass);

ALTER TABLE ONLY public.porte_sessions ALTER COLUMN id SET DEFAULT nextval('public.porte_sessions_id_seq'::regclass);

ALTER TABLE ONLY public.sites ALTER COLUMN id SET DEFAULT nextval('public.sites_id_seq'::regclass);

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);

ALTER TABLE ONLY public.visitor_sessions ALTER COLUMN id SET DEFAULT nextval('public.visitor_sessions_id_seq'::regclass);

ALTER TABLE ONLY public.webhooks ALTER COLUMN id SET DEFAULT nextval('public.webhooks_id_seq'::regclass);

ALTER TABLE ONLY public.workspace_members ALTER COLUMN id SET DEFAULT nextval('public.workspace_members_id_seq'::regclass);

ALTER TABLE ONLY public.workspaces ALTER COLUMN id SET DEFAULT nextval('public.workspaces_id_seq'::regclass);

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.custom_events
    ADD CONSTRAINT custom_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.directory_deletion_state
    ADD CONSTRAINT directory_deletion_state_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.goals
    ADD CONSTRAINT goals_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.mcp_actions
    ADD CONSTRAINT mcp_actions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.pageviews
    ADD CONSTRAINT pageviews_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.porte_identities
    ADD CONSTRAINT porte_identities_pkey PRIMARY KEY (provider, subject);

ALTER TABLE ONLY public.porte_login_codes
    ADD CONSTRAINT porte_login_codes_pkey PRIMARY KEY (code_hash);

ALTER TABLE ONLY public.porte_magic_links
    ADD CONSTRAINT porte_magic_links_pkey PRIMARY KEY (token_hash);

ALTER TABLE ONLY public.porte_sessions
    ADD CONSTRAINT porte_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.porte_sessions
    ADD CONSTRAINT porte_sessions_token_hash_key UNIQUE (token_hash);

ALTER TABLE ONLY public.sites
    ADD CONSTRAINT sites_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.visitor_sessions
    ADD CONSTRAINT visitor_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT workspace_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT workspaces_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX idx_api_keys_key_hash ON public.api_keys USING btree (key_hash);

CREATE INDEX idx_api_keys_site_id ON public.api_keys USING btree (site_id);

CREATE INDEX idx_api_keys_user_id ON public.api_keys USING btree (user_id);

CREATE INDEX idx_api_keys_workspace_id ON public.api_keys USING btree (workspace_id);

CREATE INDEX idx_custom_events_name ON public.custom_events USING btree (name);

CREATE INDEX idx_event_site_created ON public.custom_events USING btree (site_id, created_at);

CREATE INDEX idx_goals_owner_id ON public.goals USING btree (owner_id);

CREATE INDEX idx_goals_site_id ON public.goals USING btree (site_id);

CREATE INDEX idx_pageviews_created_at ON public.pageviews USING btree (created_at);

CREATE INDEX idx_pageviews_path ON public.pageviews USING btree (path);

CREATE INDEX idx_pageviews_site_id ON public.pageviews USING btree (site_id);

CREATE INDEX idx_pageviews_visitor_id ON public.pageviews USING btree (visitor_id);

CREATE INDEX idx_site_created ON public.pageviews USING btree (site_id, created_at);

CREATE UNIQUE INDEX idx_sites_domain ON public.sites USING btree (domain);

CREATE INDEX idx_sites_owner_id ON public.sites USING btree (owner_id);

CREATE UNIQUE INDEX idx_sites_share_token ON public.sites USING btree (share_token);

CREATE INDEX idx_sites_workspace_id ON public.sites USING btree (workspace_id);

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);

CREATE UNIQUE INDEX idx_users_o_id_c_subject ON public.users USING btree (oidc_subject);

CREATE INDEX idx_visitor_sessions_site_id ON public.visitor_sessions USING btree (site_id);

CREATE INDEX idx_webhooks_owner_id ON public.webhooks USING btree (owner_id);

CREATE INDEX idx_webhooks_site_id ON public.webhooks USING btree (site_id);

CREATE INDEX idx_webhooks_workspace_id ON public.webhooks USING btree (workspace_id);

CREATE INDEX idx_workspace_members_user_id ON public.workspace_members USING btree (user_id);

CREATE INDEX idx_workspace_members_workspace_id ON public.workspace_members USING btree (workspace_id);

CREATE UNIQUE INDEX idx_workspaces_registre_space_id ON public.workspaces USING btree (registre_space_id);

CREATE UNIQUE INDEX idx_ws_user ON public.workspace_members USING btree (workspace_id, user_id);

CREATE INDEX mcp_actions_open_idx ON public.mcp_actions USING btree (user_id, created_at DESC) WHERE (resolved_at IS NULL);

CREATE INDEX porte_identities_user_idx ON public.porte_identities USING btree (user_id);

CREATE INDEX porte_magic_links_user_idx ON public.porte_magic_links USING btree (user_id, created_at);

CREATE INDEX porte_sessions_expiry_idx ON public.porte_sessions USING btree (expires_at) WHERE (expires_at IS NOT NULL);

CREATE INDEX porte_sessions_user_idx ON public.porte_sessions USING btree (user_id);

ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT fk_workspace_members_user FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT fk_workspace_members_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);

ALTER TABLE ONLY public.porte_identities
    ADD CONSTRAINT porte_identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.porte_login_codes
    ADD CONSTRAINT porte_login_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.porte_magic_links
    ADD CONSTRAINT porte_magic_links_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.porte_sessions
    ADD CONSTRAINT porte_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
