-- ============================================================
-- Migration 002: Associer permissions aux rôles système
-- ============================================================

-- ── Admin → toutes les permissions ───────────────────────────
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;


-- ── Ops → rôle opérationnel complet ──────────────────────────
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON

    -- Operations
    (p.resource = 'run' AND p.action IN ('create','start','pause','stop','read'))
 OR (p.resource = 'release')
 OR (p.resource = 'ingestion')

    -- Configuration (partielle)
 OR (p.resource = 'plan' AND p.action IN ('create','update','read'))
 OR (p.resource = 'attribute' AND p.action IN ('read'))
 OR (p.resource = 'simulator' AND p.action = 'execute')

    -- Analytics
 OR (p.resource = 'dashboard' AND p.action IN ('read','export'))
 OR (p.resource = 'audit' AND p.action = 'read')

    -- Legacy utile
 OR (p.resource = 'notification' AND p.action IN ('read','write'))
 OR (p.resource = 'metrics' AND p.action = 'read')

WHERE r.name = 'ops'
ON CONFLICT DO NOTHING;


-- ── Viewer → lecture seule (read-only) ───────────────────────
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON

    (p.action = 'read' AND p.resource IN (
        'run',
        'plan',
        'attribute',
        'ingestion',
        'dashboard',
        'audit',
        'metrics',
        'notification'
    ))

WHERE r.name = 'viewer'
ON CONFLICT DO NOTHING;


-- ── (OPTIONNEL MAIS RECOMMANDÉ) Analyst ──────────────────────
-- Focus analytics + replay
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON

    (p.resource = 'audit' AND p.action IN ('read','replay'))
 OR (p.resource = 'dashboard' AND p.action IN ('read','export'))
 OR (p.resource = 'metrics' AND p.action = 'read')

WHERE r.name = 'analyst'
ON CONFLICT DO NOTHING;


-- ── (OPTIONNEL) Config Manager ───────────────────────────────
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON

    (p.resource = 'plan')
 OR (p.resource = 'attribute')
 OR (p.resource = 'settings')
 OR (p.resource = 'simulator')

WHERE r.name = 'config_manager'
ON CONFLICT DO NOTHING;


-- ── Assigner rôle admin à l'utilisateur admin seed ───────────
INSERT INTO user_roles (user_id, role_id, assigned_by)
SELECT u.id, r.id, u.id
FROM users u
JOIN roles r ON r.name = 'admin'
WHERE u.identifier = 'admin'
ON CONFLICT DO NOTHING;


SELECT 'Migration 002 appliquée.' AS status;