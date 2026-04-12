-- ============================================================
-- Migration: 001_initial_schema
-- Description: Create users table with RBAC, soft delete, JSONB metadata
-- ============================================================

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ── users ────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255)    NOT NULL,
    identifier  VARCHAR(50)    NOT NULL,
    password    VARCHAR(255)    NOT NULL,
    first_name  VARCHAR(100)    NOT NULL,
    last_name   VARCHAR(100)    NOT NULL,
    role        VARCHAR(20)     NOT NULL DEFAULT 'viewer'
                    CHECK (role IN ('admin', 'ops', 'viewer')),
    is_active   BOOLEAN         NOT NULL DEFAULT TRUE,
    metadata    JSONB,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ     -- soft delete (NULL = not deleted)
);

-- Unique identifier among non-deleted users
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_identifier_active
    ON users (identifier)
    WHERE deleted_at IS NULL;

-- Fast lookup by identifier (authentication)
CREATE INDEX IF NOT EXISTS idx_users_identifier
    ON users (identifier);

-- Soft-delete filter (GORM uses this automatically)
CREATE INDEX IF NOT EXISTS idx_users_deleted_at
    ON users (deleted_at);

-- Role-based queries
CREATE INDEX IF NOT EXISTS idx_users_role
    ON users (role)
    WHERE deleted_at IS NULL;

-- JSONB index for flexible metadata queries
CREATE INDEX IF NOT EXISTS idx_users_metadata
    ON users USING GIN (metadata);

-- Auto-update updated_at on row change
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── Seed: default admin user ─────────────────────────────────────────────────
-- Password: Admin@1234 (bcrypt, cost=10) — CHANGE IN PRODUCTION
INSERT INTO users (id, identifier, email, password, first_name, last_name, role, is_active)
VALUES (
    gen_random_uuid(),
    'admin',
    'admin@ops-server.local',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'Super',
    'Admin',
    'admin',
    TRUE
) ON CONFLICT DO NOTHING;

-- ── roles ─────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS roles (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(50)  NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description  TEXT,
    is_system    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_name ON roles (name) WHERE deleted_at IS NULL;

-- ── permissions ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource    VARCHAR(100) NOT NULL,
    action      VARCHAR(50)  NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_slug ON permissions (resource, action) WHERE deleted_at IS NULL;

-- ── role_permissions (many2many) ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- ── user_roles (many2many) ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS user_roles (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);
CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles (user_id);

-- ── notifications ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        VARCHAR(20)  NOT NULL CHECK (type IN ('email','push','in_app','sms')),
    status      VARCHAR(20)  NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','sent','failed','read')),
    title       VARCHAR(255) NOT NULL,
    body        TEXT         NOT NULL,
    payload     JSONB,
    read_at     TIMESTAMPTZ,
    sent_at     TIMESTAMPTZ,
    retry_count INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_notifications_user  ON notifications (user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications (status);

-- ── metrics ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS metrics (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(200) NOT NULL,
    type        VARCHAR(20)  NOT NULL CHECK (type IN ('counter','gauge','histogram')),
    value       DOUBLE PRECISION NOT NULL,
    labels      JSONB,
    period      VARCHAR(20),
    recorded_at TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_metrics_name        ON metrics (name);
CREATE INDEX IF NOT EXISTS idx_metrics_recorded_at ON metrics (recorded_at DESC);

-- ── events ────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(200) NOT NULL,
    severity    VARCHAR(20)  NOT NULL CHECK (severity IN ('info','warning','critical')),
    source      VARCHAR(100),
    user_id     UUID REFERENCES users(id),
    payload     JSONB,
    occurred_at TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_events_severity    ON events (severity);
CREATE INDEX IF NOT EXISTS idx_events_occurred_at ON events (occurred_at DESC);

-- ── audit_trails ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS audit_trails (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        REFERENCES users(id),
    action      VARCHAR(20) NOT NULL CHECK (action IN ('CREATE','UPDATE','DELETE','READ','LOGIN','LOGOUT')),
    resource    VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100),
    old_values  JSONB,
    new_values  JSONB,
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    request_id  VARCHAR(100),
    status_code INT         NOT NULL DEFAULT 200,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_audit_resource   ON audit_trails (resource);
CREATE INDEX IF NOT EXISTS idx_audit_user       ON audit_trails (user_id);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_trails (created_at DESC);

-- ── logs ──────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS logs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level      VARCHAR(10)  NOT NULL CHECK (level IN ('debug','info','warning','error','fatal')),
    message    TEXT         NOT NULL,
    service    VARCHAR(100),
    trace_id   VARCHAR(100),
    request_id VARCHAR(100),
    fields     JSONB,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_logs_level      ON logs (level);
CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_logs_trace_id   ON logs (trace_id) WHERE trace_id IS NOT NULL;

-- ── Seed roles système ────────────────────────────────────────────────────────
INSERT INTO roles (name, display_name, description, is_system) VALUES
  ('admin',   'Administrateur', 'Accès complet',        TRUE),
  ('ops', 'Ops',        'Gestion opérationnelle', TRUE),
  ('viewer',  'Observateur',    'Lecture seule',          TRUE)
ON CONFLICT DO NOTHING;

-- ── Seed permissions (granular RBAC) ──────────────────────────────────────────
INSERT INTO permissions (resource, action, description) VALUES

  -- ── Operations ─────────────────────────────────────────────────────────────
  ('run',        'create', 'Créer un run de sampling'),
  ('run',        'start',  'Démarrer un run de sampling'),
  ('run',        'pause',  'Mettre en pause un run de sampling'),
  ('run',        'stop',   'Arrêter un run de sampling'),
  ('run',        'read',   'Consulter les runs de sampling'),

  ('release',    'force',  'Forcer une release'),
  ('release',    'manage_targets', 'Gérer les cibles de publication'),

  ('ingestion',  'read',   'Consulter le pipeline et la santé des sources'),
  ('ingestion',  'manual_upload', 'Uploader des données manuellement (Excel/JSON)'),


  -- ── Configuration ─────────────────────────────────────────────────────────
  ('plan',       'create', 'Créer un plan de sampling'),
  ('plan',       'update', 'Modifier un plan de sampling'),
  ('plan',       'read',   'Consulter les plans de sampling'),
  ('plan',       'delete', 'Supprimer un plan de sampling'),

  ('settings',   'update', 'Modifier la configuration globale du système'),
  ('settings',   'read',   'Consulter la configuration globale'),

  ('attribute',  'create', 'Créer des attributs de schéma'),
  ('attribute',  'update', 'Modifier des attributs de schéma'),
  ('attribute',  'delete', 'Supprimer des attributs de schéma'),
  ('attribute',  'read',   'Consulter les attributs de schéma'),

  ('simulator',  'execute', 'Exécuter des simulations de sampling'),


  -- ── Administration ────────────────────────────────────────────────────────
  ('user',       'create', 'Créer des utilisateurs'),
  ('user',       'read',   'Lire les utilisateurs'),
  ('user',       'update', 'Modifier les utilisateurs'),
  ('user',       'delete', 'Supprimer les utilisateurs'),

  ('role',       'create', 'Créer des rôles'),
  ('role',       'read',   'Consulter les rôles'),
  ('role',       'update', 'Modifier les rôles'),
  ('role',       'delete', 'Supprimer les rôles'),
  ('role',       'assign', 'Assigner des rôles aux utilisateurs'),

  ('system_logs','read',   'Accéder aux logs système bas niveau'),


  -- ── Audit & Analytics ─────────────────────────────────────────────────────
  ('audit',      'read',   'Consulter les logs d’audit'),
  ('audit',      'replay', 'Rejouer les décisions de sampling'),

  ('dashboard',  'read',   'Consulter les dashboards temps réel'),
  ('dashboard',  'export', 'Exporter les données dashboards (CSV/Excel)'),


  -- ── Legacy / Generic (optionnel mais utile) ───────────────────────────────
  ('notification','read',   'Lire les notifications'),
  ('notification','write',  'Envoyer des notifications'),

  ('metrics',     'read',   'Consulter les métriques'),


  -- ── Super Admin ───────────────────────────────────────────────────────────
  ('*',          '*',      'Accès total')

ON CONFLICT DO NOTHING;


-- ── configs (dynamic system configuration) ───────────────────────────────────
-- ── configs ──────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS configs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity      VARCHAR(100) NOT NULL,
    key         VARCHAR(100) NOT NULL,
    data        JSONB        NOT NULL,
    version     INT          NOT NULL DEFAULT 1,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_by  UUID REFERENCES users(id),
    updated_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Unicité par entity + key (non supprimé)
CREATE UNIQUE INDEX IF NOT EXISTS idx_configs_entity_key_active
    ON configs (entity, key)
    WHERE deleted_at IS NULL;

-- Lookup rapide (CRITIQUE pour runtime)
CREATE INDEX IF NOT EXISTS idx_configs_entity_key
    ON configs (entity, key)
    WHERE is_active = TRUE AND deleted_at IS NULL;

-- Soft delete
CREATE INDEX IF NOT EXISTS idx_configs_deleted_at
    ON configs (deleted_at);

-- JSONB queries (si tu filtres dedans)
CREATE INDEX IF NOT EXISTS idx_configs_data
    ON configs USING GIN (data);


DROP TRIGGER IF EXISTS trigger_configs_updated_at ON configs;

CREATE TRIGGER trigger_configs_updated_at
    BEFORE UPDATE ON configs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

    -- ── Seed: LDAP configuration ────────────────────────────────────────────────
INSERT INTO configs (entity, key, data, version, is_active)
VALUES (
    'ldap',
    'default',
    '{
        "host": "ldap.example.com",
        "port": 389,
        "baseDN": "dc=example,dc=com",
        "bindDN": "cn=admin,dc=example,dc=com",
        "password": "admin123",
        "userFilter": "(uid=%s)",
        "attributes": {
            "username": "uid",
            "email": "mail",
            "firstName": "givenName",
            "lastName": "sn"
        },
        "tls": false,
        "timeout": 5
    }'::jsonb,
    1,
    TRUE
)
ON CONFLICT DO NOTHING;