-- ============================================================
-- Migration 002: Associer permissions aux rôles système
-- ============================================================

-- Admin → toutes les permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Ops → user:read, user:write, notification:read, notification:write, metrics:read
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON (p.resource = 'user'         AND p.action IN ('read','write'))
                   OR (p.resource = 'notification'  AND p.action IN ('read','write'))
                   OR (p.resource = 'metrics'       AND p.action  = 'read')
WHERE r.name = 'ops'
ON CONFLICT DO NOTHING;

-- viewer → user:read, notification:read
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON (p.resource = 'user'        AND p.action = 'read')
                   OR (p.resource = 'notification' AND p.action = 'read')
WHERE r.name = 'viewer'
ON CONFLICT DO NOTHING;

-- Assigner rôle admin à l'utilisateur admin seed
INSERT INTO user_roles (user_id, role_id, assigned_by)
SELECT u.id, r.id, u.id
FROM users u, roles r
WHERE u.identifier = 'admin'
  AND r.name  = 'admin'
ON CONFLICT DO NOTHING;

SELECT 'Migration 002 appliquée.' AS status;
