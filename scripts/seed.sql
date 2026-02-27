-- =============================================================================
-- Seed Data: Authentication & Authorization System
-- =============================================================================
-- Tables covered:
--   users, oauth_accounts, oauth_states, sessions,
--   roles, permissions, role_permissions, user_roles
--
-- Data reflects a realistic SaaS platform ("Nexus Platform") with:
--   - Engineering, Marketing, Finance, Legal, Support departments
--   - Multiple roles: admin, manager, editor, analyst, viewer, support_agent
--   - Resources: article, report, user, billing, ticket, audit_log, dashboard
--   - Scoped roles (organization-level) and global roles
-- =============================================================================

BEGIN;

-- =============================================================================
-- EXTENSIONS
-- =============================================================================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- gen_random_uuid()

-- =============================================================================
-- TRUNCATE (safe re-run)
-- =============================================================================
TRUNCATE user_roles, role_permissions, sessions, oauth_accounts, oauth_states,
         permissions, roles, users
RESTART IDENTITY CASCADE;

-- =============================================================================
-- 1. ROLES
-- =============================================================================
INSERT INTO roles (id, name, description) VALUES
  ('00000001-0000-0000-0000-000000000001', 'admin',         'Full system access including user management and billing'),
  ('00000001-0000-0000-0000-000000000002', 'manager',       'Manage team members and approve content within their department'),
  ('00000001-0000-0000-0000-000000000003', 'editor',        'Create, update, and publish content; cannot delete or manage users'),
  ('00000001-0000-0000-0000-000000000004', 'analyst',       'Read access to reports and dashboards; can export data'),
  ('00000001-0000-0000-0000-000000000005', 'viewer',        'Read-only access to published content and dashboards'),
  ('00000001-0000-0000-0000-000000000006', 'support_agent', 'Access to support tickets and limited user profile reads'),
  ('00000001-0000-0000-0000-000000000007', 'billing_admin', 'Full access to billing, invoices, and subscription management'),
  ('00000001-0000-0000-0000-000000000008', 'auditor',       'Read-only access to audit logs and compliance reports');

-- =============================================================================
-- 2. PERMISSIONS  (resource × action)
-- =============================================================================
INSERT INTO permissions (id, resource, action) VALUES
  -- article
  ('00000002-0000-0000-0000-000000000001', 'article',   'create'),
  ('00000002-0000-0000-0000-000000000002', 'article',   'read'),
  ('00000002-0000-0000-0000-000000000003', 'article',   'update'),
  ('00000002-0000-0000-0000-000000000004', 'article',   'delete'),
  ('00000002-0000-0000-0000-000000000005', 'article',   'publish'),
  ('00000002-0000-0000-0000-000000000006', 'article',   'archive'),
  -- report
  ('00000002-0000-0000-0000-000000000007', 'report',    'create'),
  ('00000002-0000-0000-0000-000000000008', 'report',    'read'),
  ('00000002-0000-0000-0000-000000000009', 'report',    'update'),
  ('00000002-0000-0000-0000-000000000010', 'report',    'delete'),
  ('00000002-0000-0000-0000-000000000011', 'report',    'export'),
  -- user
  ('00000002-0000-0000-0000-000000000012', 'user',      'create'),
  ('00000002-0000-0000-0000-000000000013', 'user',      'read'),
  ('00000002-0000-0000-0000-000000000014', 'user',      'update'),
  ('00000002-0000-0000-0000-000000000015', 'user',      'delete'),
  ('00000002-0000-0000-0000-000000000016', 'user',      'impersonate'),
  -- billing
  ('00000002-0000-0000-0000-000000000017', 'billing',   'read'),
  ('00000002-0000-0000-0000-000000000018', 'billing',   'update'),
  ('00000002-0000-0000-0000-000000000019', 'billing',   'export'),
  ('00000002-0000-0000-0000-000000000020', 'billing',   'refund'),
  -- ticket
  ('00000002-0000-0000-0000-000000000021', 'ticket',    'create'),
  ('00000002-0000-0000-0000-000000000022', 'ticket',    'read'),
  ('00000002-0000-0000-0000-000000000023', 'ticket',    'update'),
  ('00000002-0000-0000-0000-000000000024', 'ticket',    'delete'),
  ('00000002-0000-0000-0000-000000000025', 'ticket',    'assign'),
  ('00000002-0000-0000-0000-000000000026', 'ticket',    'close'),
  -- audit_log
  ('00000002-0000-0000-0000-000000000027', 'audit_log', 'read'),
  ('00000002-0000-0000-0000-000000000028', 'audit_log', 'export'),
  -- dashboard
  ('00000002-0000-0000-0000-000000000029', 'dashboard', 'read'),
  ('00000002-0000-0000-0000-000000000030', 'dashboard', 'create'),
  ('00000002-0000-0000-0000-000000000031', 'dashboard', 'update'),
  ('00000002-0000-0000-0000-000000000032', 'dashboard', 'delete'),
  ('00000002-0000-0000-0000-000000000033', 'dashboard', 'share');

-- =============================================================================
-- 3. ROLE_PERMISSIONS
-- =============================================================================

-- admin: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000001-0000-0000-0000-000000000001', id FROM permissions;

-- manager: all except impersonate, refund, audit_log export
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000001-0000-0000-0000-000000000002', id FROM permissions
WHERE (resource, action) NOT IN (
  ('user',      'impersonate'),
  ('billing',   'refund'),
  ('audit_log', 'export')
);

-- editor: article (all), report (create/read/update), dashboard (read/share), ticket (create/read)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000001-0000-0000-0000-000000000003', id FROM permissions
WHERE (resource, action) IN (
  ('article',   'create'),
  ('article',   'read'),
  ('article',   'update'),
  ('article',   'publish'),
  ('article',   'archive'),
  ('report',    'create'),
  ('report',    'read'),
  ('report',    'update'),
  ('dashboard', 'read'),
  ('dashboard', 'share'),
  ('ticket',    'create'),
  ('ticket',    'read')
);

-- analyst: report (all), dashboard (all), article (read), billing (read/export)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000001-0000-0000-0000-000000000004', id FROM permissions
WHERE (resource, action) IN (
  ('report',    'create'),
  ('report',    'read'),
  ('report',    'update'),
  ('report',    'delete'),
  ('report',    'export'),
  ('dashboard', 'read'),
  ('dashboard', 'create'),
  ('dashboard', 'update'),
  ('dashboard', 'delete'),
  ('dashboard', 'share'),
  ('article',   'read'),
  ('billing',   'read'),
  ('billing',   'export')
);

-- viewer: read-only on article, report, dashboard
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000001-0000-0000-0000-000000000005', id FROM permissions
WHERE (resource, action) IN (
  ('article',   'read'),
  ('report',    'read'),
  ('dashboard', 'read')
);

-- support_agent: ticket (all), user (read/update), article (read)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000001-0000-0000-0000-000000000006', id FROM permissions
WHERE (resource, action) IN (
  ('ticket',  'create'),
  ('ticket',  'read'),
  ('ticket',  'update'),
  ('ticket',  'assign'),
  ('ticket',  'close'),
  ('user',    'read'),
  ('user',    'update'),
  ('article', 'read')
);

-- billing_admin: billing (all), report (read/export), user (read)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000001-0000-0000-0000-000000000007', id FROM permissions
WHERE (resource, action) IN (
  ('billing', 'read'),
  ('billing', 'update'),
  ('billing', 'export'),
  ('billing', 'refund'),
  ('report',  'read'),
  ('report',  'export'),
  ('user',    'read')
);

-- auditor: audit_log (read/export), report (read/export), user (read)
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000001-0000-0000-0000-000000000008', id FROM permissions
WHERE (resource, action) IN (
  ('audit_log', 'read'),
  ('audit_log', 'export'),
  ('report',    'read'),
  ('report',    'export'),
  ('user',      'read')
);

-- =============================================================================
-- 4. USERS  (100 users across departments)
-- =============================================================================
INSERT INTO users (id, email, display_name, avatar_url, is_active, created_at, updated_at) VALUES

-- ── System / Admins ───────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000001', 'platform.admin@nexus.io',          'Platform Admin',          'https://avatars.nexus.io/platform-admin',    true,  now() - INTERVAL '730 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000002', 'devops.lead@nexus.io',             'Marcus Webb',             'https://avatars.nexus.io/marcus-webb',       true,  now() - INTERVAL '700 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000003', 'security.officer@nexus.io',        'Priya Nair',              'https://avatars.nexus.io/priya-nair',        true,  now() - INTERVAL '680 days', now() - INTERVAL '3 days'),

-- ── Engineering ───────────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000004', 'jordan.hayes@nexus.io',            'Jordan Hayes',            'https://avatars.nexus.io/jordan-hayes',      true,  now() - INTERVAL '650 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000005', 'sam.okafor@nexus.io',              'Sam Okafor',              'https://avatars.nexus.io/sam-okafor',        true,  now() - INTERVAL '620 days', now() - INTERVAL '4 days'),
('10000000-0000-0000-0000-000000000006', 'elena.voronova@nexus.io',          'Elena Voronova',          'https://avatars.nexus.io/elena-voronova',    true,  now() - INTERVAL '610 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000007', 'rafael.santos@nexus.io',           'Rafael Santos',           'https://avatars.nexus.io/rafael-santos',     true,  now() - INTERVAL '590 days', now() - INTERVAL '5 days'),
('10000000-0000-0000-0000-000000000008', 'mei.zhang@nexus.io',               'Mei Zhang',               'https://avatars.nexus.io/mei-zhang',         true,  now() - INTERVAL '580 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000009', 'tobias.mueller@nexus.io',          'Tobias Müller',           'https://avatars.nexus.io/tobias-mueller',    true,  now() - INTERVAL '560 days', now() - INTERVAL '6 days'),
('10000000-0000-0000-0000-000000000010', 'aisha.diallo@nexus.io',            'Aisha Diallo',            'https://avatars.nexus.io/aisha-diallo',      true,  now() - INTERVAL '540 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000011', 'luca.ferretti@nexus.io',           'Luca Ferretti',           'https://avatars.nexus.io/luca-ferretti',     true,  now() - INTERVAL '520 days', now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000012', 'nina.petersen@nexus.io',           'Nina Petersen',           'https://avatars.nexus.io/nina-petersen',     true,  now() - INTERVAL '510 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000013', 'kwame.asante@nexus.io',            'Kwame Asante',            'https://avatars.nexus.io/kwame-asante',      true,  now() - INTERVAL '500 days', now() - INTERVAL '7 days'),
('10000000-0000-0000-0000-000000000014', 'hana.yoshida@nexus.io',            'Hana Yoshida',            'https://avatars.nexus.io/hana-yoshida',      true,  now() - INTERVAL '490 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000015', 'omar.khalil@nexus.io',             'Omar Khalil',             'https://avatars.nexus.io/omar-khalil',       true,  now() - INTERVAL '480 days', now() - INTERVAL '4 days'),

-- ── Product ───────────────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000016', 'claire.beaumont@nexus.io',         'Claire Beaumont',         'https://avatars.nexus.io/claire-beaumont',   true,  now() - INTERVAL '470 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000017', 'derek.osei@nexus.io',              'Derek Osei',              'https://avatars.nexus.io/derek-osei',        true,  now() - INTERVAL '460 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000018', 'fatima.al-rashid@nexus.io',        'Fatima Al-Rashid',        'https://avatars.nexus.io/fatima-alrashid',   true,  now() - INTERVAL '450 days', now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000019', 'george.papadopoulos@nexus.io',     'George Papadopoulos',     'https://avatars.nexus.io/george-papa',       true,  now() - INTERVAL '440 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000020', 'ingrid.lindqvist@nexus.io',        'Ingrid Lindqvist',        'https://avatars.nexus.io/ingrid-lindqvist',  true,  now() - INTERVAL '430 days', now() - INTERVAL '5 days'),

-- ── Marketing ─────────────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000021', 'james.okonkwo@nexus.io',           'James Okonkwo',           'https://avatars.nexus.io/james-okonkwo',     true,  now() - INTERVAL '420 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000022', 'katarina.novak@nexus.io',          'Katarina Novák',          'https://avatars.nexus.io/katarina-novak',    true,  now() - INTERVAL '410 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000023', 'leila.hosseini@nexus.io',          'Leila Hosseini',          'https://avatars.nexus.io/leila-hosseini',    true,  now() - INTERVAL '400 days', now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000024', 'michael.brennan@nexus.io',         'Michael Brennan',         'https://avatars.nexus.io/michael-brennan',   true,  now() - INTERVAL '390 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000025', 'nadia.kowalski@nexus.io',          'Nadia Kowalski',          'https://avatars.nexus.io/nadia-kowalski',    true,  now() - INTERVAL '380 days', now() - INTERVAL '4 days'),
('10000000-0000-0000-0000-000000000026', 'oliver.nakamura@nexus.io',         'Oliver Nakamura',         'https://avatars.nexus.io/oliver-nakamura',   true,  now() - INTERVAL '370 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000027', 'petra.svoboda@nexus.io',           'Petra Svoboda',           'https://avatars.nexus.io/petra-svoboda',     true,  now() - INTERVAL '360 days', now() - INTERVAL '6 days'),
('10000000-0000-0000-0000-000000000028', 'quincy.adeyemi@nexus.io',          'Quincy Adeyemi',          'https://avatars.nexus.io/quincy-adeyemi',    true,  now() - INTERVAL '350 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000029', 'rosa.villanueva@nexus.io',         'Rosa Villanueva',         'https://avatars.nexus.io/rosa-villanueva',   true,  now() - INTERVAL '340 days', now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000030', 'stefan.bergmann@nexus.io',         'Stefan Bergmann',         'https://avatars.nexus.io/stefan-bergmann',   true,  now() - INTERVAL '330 days', now() - INTERVAL '1 day'),

-- ── Finance ───────────────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000031', 'tanya.morrison@nexus.io',          'Tanya Morrison',          'https://avatars.nexus.io/tanya-morrison',    true,  now() - INTERVAL '320 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000032', 'umar.ibrahim@nexus.io',            'Umar Ibrahim',            'https://avatars.nexus.io/umar-ibrahim',      true,  now() - INTERVAL '310 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000033', 'valeria.greco@nexus.io',           'Valeria Greco',           'https://avatars.nexus.io/valeria-greco',     true,  now() - INTERVAL '300 days', now() - INTERVAL '4 days'),
('10000000-0000-0000-0000-000000000034', 'william.chukwu@nexus.io',          'William Chukwu',          'https://avatars.nexus.io/william-chukwu',    true,  now() - INTERVAL '290 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000035', 'xiomara.delgado@nexus.io',         'Xiomara Delgado',         'https://avatars.nexus.io/xiomara-delgado',   true,  now() - INTERVAL '280 days', now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000036', 'yusuf.erdogan@nexus.io',           'Yusuf Erdoğan',           'https://avatars.nexus.io/yusuf-erdogan',     true,  now() - INTERVAL '270 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000037', 'zoe.karamanlis@nexus.io',          'Zoe Karamanlis',          'https://avatars.nexus.io/zoe-karamanlis',    true,  now() - INTERVAL '260 days', now() - INTERVAL '5 days'),

-- ── Legal & Compliance ────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000038', 'aaron.fitzgerald@nexus.io',        'Aaron Fitzgerald',        'https://avatars.nexus.io/aaron-fitzgerald',  true,  now() - INTERVAL '250 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000039', 'beatrice.fontaine@nexus.io',       'Beatrice Fontaine',       'https://avatars.nexus.io/beatrice-fontaine', true,  now() - INTERVAL '240 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000040', 'carlos.reyes@nexus.io',            'Carlos Reyes',            'https://avatars.nexus.io/carlos-reyes',      true,  now() - INTERVAL '230 days', now() - INTERVAL '3 days'),

-- ── Support ───────────────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000041', 'diana.kruger@nexus.io',            'Diana Kruger',            'https://avatars.nexus.io/diana-kruger',      true,  now() - INTERVAL '220 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000042', 'ethan.obi@nexus.io',               'Ethan Obi',               'https://avatars.nexus.io/ethan-obi',         true,  now() - INTERVAL '215 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000043', 'fiona.mcallister@nexus.io',        'Fiona McAllister',        'https://avatars.nexus.io/fiona-mcallister',  true,  now() - INTERVAL '210 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000044', 'gabriel.moreau@nexus.io',          'Gabriel Moreau',          'https://avatars.nexus.io/gabriel-moreau',    true,  now() - INTERVAL '205 days', now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000045', 'helena.svensson@nexus.io',         'Helena Svensson',         'https://avatars.nexus.io/helena-svensson',   true,  now() - INTERVAL '200 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000046', 'ibrahim.toure@nexus.io',           'Ibrahim Touré',           'https://avatars.nexus.io/ibrahim-toure',     true,  now() - INTERVAL '195 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000047', 'julia.santos@nexus.io',            'Julia Santos',            'https://avatars.nexus.io/julia-santos',      true,  now() - INTERVAL '190 days', now() - INTERVAL '4 days'),
('10000000-0000-0000-0000-000000000048', 'kevin.oduya@nexus.io',             'Kevin Oduya',             'https://avatars.nexus.io/kevin-oduya',       true,  now() - INTERVAL '185 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000049', 'lisa.hartmann@nexus.io',           'Lisa Hartmann',           'https://avatars.nexus.io/lisa-hartmann',     true,  now() - INTERVAL '180 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000050', 'mario.esposito@nexus.io',          'Mario Esposito',          'https://avatars.nexus.io/mario-esposito',    true,  now() - INTERVAL '175 days', now() - INTERVAL '3 days'),

-- ── Data / Analytics ─────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000051', 'naomi.park@nexus.io',              'Naomi Park',              'https://avatars.nexus.io/naomi-park',        true,  now() - INTERVAL '170 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000052', 'oscar.lindberg@nexus.io',          'Oscar Lindberg',          'https://avatars.nexus.io/oscar-lindberg',    true,  now() - INTERVAL '165 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000053', 'paula.ferreira@nexus.io',          'Paula Ferreira',          'https://avatars.nexus.io/paula-ferreira',    true,  now() - INTERVAL '160 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000054', 'ravi.krishnamurthy@nexus.io',      'Ravi Krishnamurthy',      'https://avatars.nexus.io/ravi-krishna',      true,  now() - INTERVAL '155 days', now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000055', 'sophie.lambert@nexus.io',          'Sophie Lambert',          'https://avatars.nexus.io/sophie-lambert',    true,  now() - INTERVAL '150 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000056', 'takeshi.yamamoto@nexus.io',        'Takeshi Yamamoto',        'https://avatars.nexus.io/takeshi-yamamoto',  true,  now() - INTERVAL '145 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000057', 'ursula.becker@nexus.io',           'Ursula Becker',           'https://avatars.nexus.io/ursula-becker',     true,  now() - INTERVAL '140 days', now() - INTERVAL '4 days'),
('10000000-0000-0000-0000-000000000058', 'victor.nwosu@nexus.io',            'Victor Nwosu',            'https://avatars.nexus.io/victor-nwosu',      true,  now() - INTERVAL '135 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000059', 'wendy.schultz@nexus.io',           'Wendy Schultz',           'https://avatars.nexus.io/wendy-schultz',     true,  now() - INTERVAL '130 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000060', 'xavier.morales@nexus.io',          'Xavier Morales',          'https://avatars.nexus.io/xavier-morales',    true,  now() - INTERVAL '125 days', now() - INTERVAL '3 days'),

-- ── Sales ────────────────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000061', 'yasmin.ali@nexus.io',              'Yasmin Ali',              'https://avatars.nexus.io/yasmin-ali',        true,  now() - INTERVAL '120 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000062', 'zachary.obinna@nexus.io',          'Zachary Obinna',          'https://avatars.nexus.io/zachary-obinna',    true,  now() - INTERVAL '118 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000063', 'alice.bertrand@nexus.io',          'Alice Bertrand',          'https://avatars.nexus.io/alice-bertrand',    true,  now() - INTERVAL '115 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000064', 'ben.ochieng@nexus.io',             'Ben Ochieng',             'https://avatars.nexus.io/ben-ochieng',       true,  now() - INTERVAL '112 days', now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000065', 'chloe.harrison@nexus.io',          'Chloe Harrison',          'https://avatars.nexus.io/chloe-harrison',    true,  now() - INTERVAL '110 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000066', 'damian.kowalczyk@nexus.io',        'Damian Kowalczyk',        'https://avatars.nexus.io/damian-kowalczyk',  true,  now() - INTERVAL '108 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000067', 'emma.johansson@nexus.io',          'Emma Johansson',          'https://avatars.nexus.io/emma-johansson',    true,  now() - INTERVAL '105 days', now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000068', 'felix.wagner@nexus.io',            'Felix Wagner',            'https://avatars.nexus.io/felix-wagner',      true,  now() - INTERVAL '102 days', now() - INTERVAL '4 days'),
('10000000-0000-0000-0000-000000000069', 'grace.afolabi@nexus.io',           'Grace Afolabi',           'https://avatars.nexus.io/grace-afolabi',     true,  now() - INTERVAL '100 days', now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000070', 'henry.bouchard@nexus.io',          'Henry Bouchard',          'https://avatars.nexus.io/henry-bouchard',    true,  now() - INTERVAL '98 days',  now() - INTERVAL '1 day'),

-- ── Design ───────────────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000071', 'isabelle.morin@nexus.io',          'Isabelle Morin',          'https://avatars.nexus.io/isabelle-morin',    true,  now() - INTERVAL '95 days',  now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000072', 'jake.odunbaku@nexus.io',           'Jake Odunbaku',           'https://avatars.nexus.io/jake-odunbaku',     true,  now() - INTERVAL '92 days',  now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000073', 'karen.steinberg@nexus.io',         'Karen Steinberg',         'https://avatars.nexus.io/karen-steinberg',   true,  now() - INTERVAL '90 days',  now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000074', 'lars.andersen@nexus.io',           'Lars Andersen',           'https://avatars.nexus.io/lars-andersen',     true,  now() - INTERVAL '87 days',  now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000075', 'mia.tanaka@nexus.io',              'Mia Tanaka',              'https://avatars.nexus.io/mia-tanaka',        true,  now() - INTERVAL '85 days',  now() - INTERVAL '1 day'),

-- ── HR ───────────────────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000076', 'noah.osei-bonsu@nexus.io',         'Noah Osei-Bonsu',         'https://avatars.nexus.io/noah-osei-bonsu',   true,  now() - INTERVAL '82 days',  now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000077', 'olivia.patel@nexus.io',            'Olivia Patel',            'https://avatars.nexus.io/olivia-patel',      true,  now() - INTERVAL '80 days',  now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000078', 'paul.meier@nexus.io',              'Paul Meier',              'https://avatars.nexus.io/paul-meier',        true,  now() - INTERVAL '78 days',  now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000079', 'quinn.adebisi@nexus.io',           'Quinn Adebisi',           'https://avatars.nexus.io/quinn-adebisi',     true,  now() - INTERVAL '75 days',  now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000080', 'rachel.dupont@nexus.io',           'Rachel Dupont',           'https://avatars.nexus.io/rachel-dupont',     true,  now() - INTERVAL '72 days',  now() - INTERVAL '2 days'),

-- ── Recent Hires ─────────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000081', 'sebastian.wirth@nexus.io',         'Sebastian Wirth',         'https://avatars.nexus.io/sebastian-wirth',   true,  now() - INTERVAL '60 days',  now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000082', 'talia.mwangi@nexus.io',            'Talia Mwangi',            'https://avatars.nexus.io/talia-mwangi',      true,  now() - INTERVAL '55 days',  now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000083', 'ulrich.hoffmann@nexus.io',         'Ulrich Hoffmann',         'https://avatars.nexus.io/ulrich-hoffmann',   true,  now() - INTERVAL '50 days',  now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000084', 'vera.antonova@nexus.io',           'Vera Antonova',           'https://avatars.nexus.io/vera-antonova',     true,  now() - INTERVAL '45 days',  now() - INTERVAL '3 days'),
('10000000-0000-0000-0000-000000000085', 'walter.nguyen@nexus.io',           'Walter Nguyen',           'https://avatars.nexus.io/walter-nguyen',     true,  now() - INTERVAL '40 days',  now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000086', 'xena.obiakor@nexus.io',            'Xena Obiakor',            'https://avatars.nexus.io/xena-obiakor',      true,  now() - INTERVAL '35 days',  now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000087', 'yannick.rousseau@nexus.io',        'Yannick Rousseau',        'https://avatars.nexus.io/yannick-rousseau',  true,  now() - INTERVAL '30 days',  now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000088', 'zara.mensah@nexus.io',             'Zara Mensah',             'https://avatars.nexus.io/zara-mensah',       true,  now() - INTERVAL '25 days',  now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000089', 'adam.kowalski@nexus.io',           'Adam Kowalski',           'https://avatars.nexus.io/adam-kowalski',     true,  now() - INTERVAL '20 days',  now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000090', 'bella.nkosi@nexus.io',             'Bella Nkosi',             'https://avatars.nexus.io/bella-nkosi',       true,  now() - INTERVAL '15 days',  now() - INTERVAL '3 days'),

-- ── Inactive / Offboarded ─────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000091', 'chris.olawale@nexus.io',           'Chris Olawale',           'https://avatars.nexus.io/chris-olawale',     false, now() - INTERVAL '400 days', now() - INTERVAL '200 days'),
('10000000-0000-0000-0000-000000000092', 'diana.fischer@nexus.io',           'Diana Fischer',           'https://avatars.nexus.io/diana-fischer',     false, now() - INTERVAL '380 days', now() - INTERVAL '180 days'),
('10000000-0000-0000-0000-000000000093', 'evan.oduola@nexus.io',             'Evan Oduola',             'https://avatars.nexus.io/evan-oduola',       false, now() - INTERVAL '360 days', now() - INTERVAL '160 days'),
('10000000-0000-0000-0000-000000000094', 'faye.leclerc@nexus.io',            'Faye Leclerc',            'https://avatars.nexus.io/faye-leclerc',      false, now() - INTERVAL '300 days', now() - INTERVAL '120 days'),
('10000000-0000-0000-0000-000000000095', 'glen.abubakar@nexus.io',           'Glen Abubakar',           'https://avatars.nexus.io/glen-abubakar',     false, now() - INTERVAL '280 days', now() - INTERVAL '100 days'),

-- ── External Contractors ─────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000096', 'hana.schmidt@contractor.io',       'Hana Schmidt',            'https://avatars.nexus.io/hana-schmidt',      true,  now() - INTERVAL '10 days',  now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000097', 'ivan.petrov@contractor.io',        'Ivan Petrov',             'https://avatars.nexus.io/ivan-petrov',       true,  now() - INTERVAL '8 days',   now() - INTERVAL '2 days'),
('10000000-0000-0000-0000-000000000098', 'jade.owusu@contractor.io',         'Jade Owusu',              'https://avatars.nexus.io/jade-owusu',        true,  now() - INTERVAL '6 days',   now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000099', 'kai.bergstrom@contractor.io',      'Kai Bergström',           'https://avatars.nexus.io/kai-bergstrom',     true,  now() - INTERVAL '4 days',   now() - INTERVAL '1 day'),
('10000000-0000-0000-0000-000000000100', 'luna.diaz@contractor.io',          'Luna Díaz',               'https://avatars.nexus.io/luna-diaz',         true,  now() - INTERVAL '2 days',   now() - INTERVAL '1 day');

-- =============================================================================
-- 5. OAUTH_ACCOUNTS  (Google provider_sub mirrors a real Google ID format)
-- =============================================================================
INSERT INTO oauth_accounts (id, user_id, provider, provider_sub, email, last_login, created_at) VALUES
('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'google', '100000000000000000001', 'platform.admin@nexus.io',      now() - INTERVAL '1 day',   now() - INTERVAL '730 days'),
('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', 'google', '100000000000000000002', 'devops.lead@nexus.io',          now() - INTERVAL '2 days',  now() - INTERVAL '700 days'),
('20000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000003', 'google', '100000000000000000003', 'security.officer@nexus.io',     now() - INTERVAL '3 days',  now() - INTERVAL '680 days'),
('20000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000004', 'google', '100000000000000000004', 'jordan.hayes@nexus.io',         now() - INTERVAL '1 day',   now() - INTERVAL '650 days'),
('20000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000005', 'google', '100000000000000000005', 'sam.okafor@nexus.io',           now() - INTERVAL '4 days',  now() - INTERVAL '620 days'),
('20000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000006', 'google', '100000000000000000006', 'elena.voronova@nexus.io',       now() - INTERVAL '2 days',  now() - INTERVAL '610 days'),
('20000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000007', 'google', '100000000000000000007', 'rafael.santos@nexus.io',        now() - INTERVAL '5 days',  now() - INTERVAL '590 days'),
('20000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000008', 'google', '100000000000000000008', 'mei.zhang@nexus.io',            now() - INTERVAL '1 day',   now() - INTERVAL '580 days'),
('20000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000009', 'google', '100000000000000000009', 'tobias.mueller@nexus.io',       now() - INTERVAL '6 days',  now() - INTERVAL '560 days'),
('20000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000010', 'google', '100000000000000000010', 'aisha.diallo@nexus.io',         now() - INTERVAL '2 days',  now() - INTERVAL '540 days'),
('20000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000011', 'google', '100000000000000000011', 'luca.ferretti@nexus.io',        now() - INTERVAL '3 days',  now() - INTERVAL '520 days'),
('20000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000012', 'google', '100000000000000000012', 'nina.petersen@nexus.io',        now() - INTERVAL '1 day',   now() - INTERVAL '510 days'),
('20000000-0000-0000-0000-000000000013', '10000000-0000-0000-0000-000000000013', 'google', '100000000000000000013', 'kwame.asante@nexus.io',         now() - INTERVAL '7 days',  now() - INTERVAL '500 days'),
('20000000-0000-0000-0000-000000000014', '10000000-0000-0000-0000-000000000014', 'google', '100000000000000000014', 'hana.yoshida@nexus.io',         now() - INTERVAL '2 days',  now() - INTERVAL '490 days'),
('20000000-0000-0000-0000-000000000015', '10000000-0000-0000-0000-000000000015', 'google', '100000000000000000015', 'omar.khalil@nexus.io',          now() - INTERVAL '4 days',  now() - INTERVAL '480 days'),
('20000000-0000-0000-0000-000000000016', '10000000-0000-0000-0000-000000000016', 'google', '100000000000000000016', 'claire.beaumont@nexus.io',      now() - INTERVAL '1 day',   now() - INTERVAL '470 days'),
('20000000-0000-0000-0000-000000000017', '10000000-0000-0000-0000-000000000017', 'google', '100000000000000000017', 'derek.osei@nexus.io',           now() - INTERVAL '2 days',  now() - INTERVAL '460 days'),
('20000000-0000-0000-0000-000000000018', '10000000-0000-0000-0000-000000000018', 'google', '100000000000000000018', 'fatima.al-rashid@nexus.io',     now() - INTERVAL '3 days',  now() - INTERVAL '450 days'),
('20000000-0000-0000-0000-000000000019', '10000000-0000-0000-0000-000000000019', 'google', '100000000000000000019', 'george.papadopoulos@nexus.io',  now() - INTERVAL '1 day',   now() - INTERVAL '440 days'),
('20000000-0000-0000-0000-000000000020', '10000000-0000-0000-0000-000000000020', 'google', '100000000000000000020', 'ingrid.lindqvist@nexus.io',     now() - INTERVAL '5 days',  now() - INTERVAL '430 days'),
('20000000-0000-0000-0000-000000000021', '10000000-0000-0000-0000-000000000021', 'google', '100000000000000000021', 'james.okonkwo@nexus.io',        now() - INTERVAL '2 days',  now() - INTERVAL '420 days'),
('20000000-0000-0000-0000-000000000022', '10000000-0000-0000-0000-000000000022', 'google', '100000000000000000022', 'katarina.novak@nexus.io',       now() - INTERVAL '1 day',   now() - INTERVAL '410 days'),
('20000000-0000-0000-0000-000000000023', '10000000-0000-0000-0000-000000000023', 'google', '100000000000000000023', 'leila.hosseini@nexus.io',       now() - INTERVAL '3 days',  now() - INTERVAL '400 days'),
('20000000-0000-0000-0000-000000000024', '10000000-0000-0000-0000-000000000024', 'google', '100000000000000000024', 'michael.brennan@nexus.io',      now() - INTERVAL '2 days',  now() - INTERVAL '390 days'),
('20000000-0000-0000-0000-000000000025', '10000000-0000-0000-0000-000000000025', 'google', '100000000000000000025', 'nadia.kowalski@nexus.io',       now() - INTERVAL '4 days',  now() - INTERVAL '380 days'),
('20000000-0000-0000-0000-000000000026', '10000000-0000-0000-0000-000000000026', 'google', '100000000000000000026', 'oliver.nakamura@nexus.io',      now() - INTERVAL '1 day',   now() - INTERVAL '370 days'),
('20000000-0000-0000-0000-000000000027', '10000000-0000-0000-0000-000000000027', 'google', '100000000000000000027', 'petra.svoboda@nexus.io',        now() - INTERVAL '6 days',  now() - INTERVAL '360 days'),
('20000000-0000-0000-0000-000000000028', '10000000-0000-0000-0000-000000000028', 'google', '100000000000000000028', 'quincy.adeyemi@nexus.io',       now() - INTERVAL '2 days',  now() - INTERVAL '350 days'),
('20000000-0000-0000-0000-000000000029', '10000000-0000-0000-0000-000000000029', 'google', '100000000000000000029', 'rosa.villanueva@nexus.io',      now() - INTERVAL '3 days',  now() - INTERVAL '340 days'),
('20000000-0000-0000-0000-000000000030', '10000000-0000-0000-0000-000000000030', 'google', '100000000000000000030', 'stefan.bergmann@nexus.io',      now() - INTERVAL '1 day',   now() - INTERVAL '330 days'),
('20000000-0000-0000-0000-000000000031', '10000000-0000-0000-0000-000000000031', 'google', '100000000000000000031', 'tanya.morrison@nexus.io',       now() - INTERVAL '2 days',  now() - INTERVAL '320 days'),
('20000000-0000-0000-0000-000000000032', '10000000-0000-0000-0000-000000000032', 'google', '100000000000000000032', 'umar.ibrahim@nexus.io',         now() - INTERVAL '1 day',   now() - INTERVAL '310 days'),
('20000000-0000-0000-0000-000000000033', '10000000-0000-0000-0000-000000000033', 'google', '100000000000000000033', 'valeria.greco@nexus.io',        now() - INTERVAL '4 days',  now() - INTERVAL '300 days'),
('20000000-0000-0000-0000-000000000034', '10000000-0000-0000-0000-000000000034', 'google', '100000000000000000034', 'william.chukwu@nexus.io',       now() - INTERVAL '2 days',  now() - INTERVAL '290 days'),
('20000000-0000-0000-0000-000000000035', '10000000-0000-0000-0000-000000000035', 'google', '100000000000000000035', 'xiomara.delgado@nexus.io',      now() - INTERVAL '3 days',  now() - INTERVAL '280 days'),
('20000000-0000-0000-0000-000000000036', '10000000-0000-0000-0000-000000000036', 'google', '100000000000000000036', 'yusuf.erdogan@nexus.io',        now() - INTERVAL '1 day',   now() - INTERVAL '270 days'),
('20000000-0000-0000-0000-000000000037', '10000000-0000-0000-0000-000000000037', 'google', '100000000000000000037', 'zoe.karamanlis@nexus.io',       now() - INTERVAL '5 days',  now() - INTERVAL '260 days'),
('20000000-0000-0000-0000-000000000038', '10000000-0000-0000-0000-000000000038', 'google', '100000000000000000038', 'aaron.fitzgerald@nexus.io',     now() - INTERVAL '2 days',  now() - INTERVAL '250 days'),
('20000000-0000-0000-0000-000000000039', '10000000-0000-0000-0000-000000000039', 'google', '100000000000000000039', 'beatrice.fontaine@nexus.io',    now() - INTERVAL '1 day',   now() - INTERVAL '240 days'),
('20000000-0000-0000-0000-000000000040', '10000000-0000-0000-0000-000000000040', 'google', '100000000000000000040', 'carlos.reyes@nexus.io',         now() - INTERVAL '3 days',  now() - INTERVAL '230 days'),
('20000000-0000-0000-0000-000000000041', '10000000-0000-0000-0000-000000000041', 'google', '100000000000000000041', 'diana.kruger@nexus.io',         now() - INTERVAL '1 day',   now() - INTERVAL '220 days'),
('20000000-0000-0000-0000-000000000042', '10000000-0000-0000-0000-000000000042', 'google', '100000000000000000042', 'ethan.obi@nexus.io',            now() - INTERVAL '2 days',  now() - INTERVAL '215 days'),
('20000000-0000-0000-0000-000000000043', '10000000-0000-0000-0000-000000000043', 'google', '100000000000000000043', 'fiona.mcallister@nexus.io',     now() - INTERVAL '1 day',   now() - INTERVAL '210 days'),
('20000000-0000-0000-0000-000000000044', '10000000-0000-0000-0000-000000000044', 'google', '100000000000000000044', 'gabriel.moreau@nexus.io',       now() - INTERVAL '3 days',  now() - INTERVAL '205 days'),
('20000000-0000-0000-0000-000000000045', '10000000-0000-0000-0000-000000000045', 'google', '100000000000000000045', 'helena.svensson@nexus.io',      now() - INTERVAL '2 days',  now() - INTERVAL '200 days'),
('20000000-0000-0000-0000-000000000046', '10000000-0000-0000-0000-000000000046', 'google', '100000000000000000046', 'ibrahim.toure@nexus.io',        now() - INTERVAL '1 day',   now() - INTERVAL '195 days'),
('20000000-0000-0000-0000-000000000047', '10000000-0000-0000-0000-000000000047', 'google', '100000000000000000047', 'julia.santos@nexus.io',         now() - INTERVAL '4 days',  now() - INTERVAL '190 days'),
('20000000-0000-0000-0000-000000000048', '10000000-0000-0000-0000-000000000048', 'google', '100000000000000000048', 'kevin.oduya@nexus.io',          now() - INTERVAL '2 days',  now() - INTERVAL '185 days'),
('20000000-0000-0000-0000-000000000049', '10000000-0000-0000-0000-000000000049', 'google', '100000000000000000049', 'lisa.hartmann@nexus.io',        now() - INTERVAL '1 day',   now() - INTERVAL '180 days'),
('20000000-0000-0000-0000-000000000050', '10000000-0000-0000-0000-000000000050', 'google', '100000000000000000050', 'mario.esposito@nexus.io',       now() - INTERVAL '3 days',  now() - INTERVAL '175 days'),
('20000000-0000-0000-0000-000000000051', '10000000-0000-0000-0000-000000000051', 'google', '100000000000000000051', 'naomi.park@nexus.io',           now() - INTERVAL '1 day',   now() - INTERVAL '170 days'),
('20000000-0000-0000-0000-000000000052', '10000000-0000-0000-0000-000000000052', 'google', '100000000000000000052', 'oscar.lindberg@nexus.io',       now() - INTERVAL '2 days',  now() - INTERVAL '165 days'),
('20000000-0000-0000-0000-000000000053', '10000000-0000-0000-0000-000000000053', 'google', '100000000000000000053', 'paula.ferreira@nexus.io',       now() - INTERVAL '1 day',   now() - INTERVAL '160 days'),
('20000000-0000-0000-0000-000000000054', '10000000-0000-0000-0000-000000000054', 'google', '100000000000000000054', 'ravi.krishnamurthy@nexus.io',   now() - INTERVAL '3 days',  now() - INTERVAL '155 days'),
('20000000-0000-0000-0000-000000000055', '10000000-0000-0000-0000-000000000055', 'google', '100000000000000000055', 'sophie.lambert@nexus.io',       now() - INTERVAL '2 days',  now() - INTERVAL '150 days'),
('20000000-0000-0000-0000-000000000056', '10000000-0000-0000-0000-000000000056', 'google', '100000000000000000056', 'takeshi.yamamoto@nexus.io',     now() - INTERVAL '1 day',   now() - INTERVAL '145 days'),
('20000000-0000-0000-0000-000000000057', '10000000-0000-0000-0000-000000000057', 'google', '100000000000000000057', 'ursula.becker@nexus.io',        now() - INTERVAL '4 days',  now() - INTERVAL '140 days'),
('20000000-0000-0000-0000-000000000058', '10000000-0000-0000-0000-000000000058', 'google', '100000000000000000058', 'victor.nwosu@nexus.io',         now() - INTERVAL '2 days',  now() - INTERVAL '135 days'),
('20000000-0000-0000-0000-000000000059', '10000000-0000-0000-0000-000000000059', 'google', '100000000000000000059', 'wendy.schultz@nexus.io',        now() - INTERVAL '1 day',   now() - INTERVAL '130 days'),
('20000000-0000-0000-0000-000000000060', '10000000-0000-0000-0000-000000000060', 'google', '100000000000000000060', 'xavier.morales@nexus.io',       now() - INTERVAL '3 days',  now() - INTERVAL '125 days'),
('20000000-0000-0000-0000-000000000061', '10000000-0000-0000-0000-000000000061', 'google', '100000000000000000061', 'yasmin.ali@nexus.io',           now() - INTERVAL '1 day',   now() - INTERVAL '120 days'),
('20000000-0000-0000-0000-000000000062', '10000000-0000-0000-0000-000000000062', 'google', '100000000000000000062', 'zachary.obinna@nexus.io',       now() - INTERVAL '2 days',  now() - INTERVAL '118 days'),
('20000000-0000-0000-0000-000000000063', '10000000-0000-0000-0000-000000000063', 'google', '100000000000000000063', 'alice.bertrand@nexus.io',       now() - INTERVAL '1 day',   now() - INTERVAL '115 days'),
('20000000-0000-0000-0000-000000000064', '10000000-0000-0000-0000-000000000064', 'google', '100000000000000000064', 'ben.ochieng@nexus.io',          now() - INTERVAL '3 days',  now() - INTERVAL '112 days'),
('20000000-0000-0000-0000-000000000065', '10000000-0000-0000-0000-000000000065', 'google', '100000000000000000065', 'chloe.harrison@nexus.io',       now() - INTERVAL '1 day',   now() - INTERVAL '110 days'),
('20000000-0000-0000-0000-000000000066', '10000000-0000-0000-0000-000000000066', 'google', '100000000000000000066', 'damian.kowalczyk@nexus.io',     now() - INTERVAL '2 days',  now() - INTERVAL '108 days'),
('20000000-0000-0000-0000-000000000067', '10000000-0000-0000-0000-000000000067', 'google', '100000000000000000067', 'emma.johansson@nexus.io',       now() - INTERVAL '1 day',   now() - INTERVAL '105 days'),
('20000000-0000-0000-0000-000000000068', '10000000-0000-0000-0000-000000000068', 'google', '100000000000000000068', 'felix.wagner@nexus.io',         now() - INTERVAL '4 days',  now() - INTERVAL '102 days'),
('20000000-0000-0000-0000-000000000069', '10000000-0000-0000-0000-000000000069', 'google', '100000000000000000069', 'grace.afolabi@nexus.io',        now() - INTERVAL '2 days',  now() - INTERVAL '100 days'),
('20000000-0000-0000-0000-000000000070', '10000000-0000-0000-0000-000000000070', 'google', '100000000000000000070', 'henry.bouchard@nexus.io',       now() - INTERVAL '1 day',   now() - INTERVAL '98 days'),
('20000000-0000-0000-0000-000000000071', '10000000-0000-0000-0000-000000000071', 'google', '100000000000000000071', 'isabelle.morin@nexus.io',       now() - INTERVAL '2 days',  now() - INTERVAL '95 days'),
('20000000-0000-0000-0000-000000000072', '10000000-0000-0000-0000-000000000072', 'google', '100000000000000000072', 'jake.odunbaku@nexus.io',        now() - INTERVAL '1 day',   now() - INTERVAL '92 days'),
('20000000-0000-0000-0000-000000000073', '10000000-0000-0000-0000-000000000073', 'google', '100000000000000000073', 'karen.steinberg@nexus.io',      now() - INTERVAL '3 days',  now() - INTERVAL '90 days'),
('20000000-0000-0000-0000-000000000074', '10000000-0000-0000-0000-000000000074', 'google', '100000000000000000074', 'lars.andersen@nexus.io',        now() - INTERVAL '2 days',  now() - INTERVAL '87 days'),
('20000000-0000-0000-0000-000000000075', '10000000-0000-0000-0000-000000000075', 'google', '100000000000000000075', 'mia.tanaka@nexus.io',           now() - INTERVAL '1 day',   now() - INTERVAL '85 days'),
('20000000-0000-0000-0000-000000000076', '10000000-0000-0000-0000-000000000076', 'google', '100000000000000000076', 'noah.osei-bonsu@nexus.io',      now() - INTERVAL '2 days',  now() - INTERVAL '82 days'),
('20000000-0000-0000-0000-000000000077', '10000000-0000-0000-0000-000000000077', 'google', '100000000000000000077', 'olivia.patel@nexus.io',         now() - INTERVAL '1 day',   now() - INTERVAL '80 days'),
('20000000-0000-0000-0000-000000000078', '10000000-0000-0000-0000-000000000078', 'google', '100000000000000000078', 'paul.meier@nexus.io',           now() - INTERVAL '3 days',  now() - INTERVAL '78 days'),
('20000000-0000-0000-0000-000000000079', '10000000-0000-0000-0000-000000000079', 'google', '100000000000000000079', 'quinn.adebisi@nexus.io',        now() - INTERVAL '1 day',   now() - INTERVAL '75 days'),
('20000000-0000-0000-0000-000000000080', '10000000-0000-0000-0000-000000000080', 'google', '100000000000000000080', 'rachel.dupont@nexus.io',        now() - INTERVAL '2 days',  now() - INTERVAL '72 days'),
('20000000-0000-0000-0000-000000000081', '10000000-0000-0000-0000-000000000081', 'google', '100000000000000000081', 'sebastian.wirth@nexus.io',      now() - INTERVAL '1 day',   now() - INTERVAL '60 days'),
('20000000-0000-0000-0000-000000000082', '10000000-0000-0000-0000-000000000082', 'google', '100000000000000000082', 'talia.mwangi@nexus.io',         now() - INTERVAL '2 days',  now() - INTERVAL '55 days'),
('20000000-0000-0000-0000-000000000083', '10000000-0000-0000-0000-000000000083', 'google', '100000000000000000083', 'ulrich.hoffmann@nexus.io',      now() - INTERVAL '1 day',   now() - INTERVAL '50 days'),
('20000000-0000-0000-0000-000000000084', '10000000-0000-0000-0000-000000000084', 'google', '100000000000000000084', 'vera.antonova@nexus.io',        now() - INTERVAL '3 days',  now() - INTERVAL '45 days'),
('20000000-0000-0000-0000-000000000085', '10000000-0000-0000-0000-000000000085', 'google', '100000000000000000085', 'walter.nguyen@nexus.io',        now() - INTERVAL '1 day',   now() - INTERVAL '40 days'),
('20000000-0000-0000-0000-000000000086', '10000000-0000-0000-0000-000000000086', 'google', '100000000000000000086', 'xena.obiakor@nexus.io',         now() - INTERVAL '2 days',  now() - INTERVAL '35 days'),
('20000000-0000-0000-0000-000000000087', '10000000-0000-0000-0000-000000000087', 'google', '100000000000000000087', 'yannick.rousseau@nexus.io',     now() - INTERVAL '1 day',   now() - INTERVAL '30 days'),
('20000000-0000-0000-0000-000000000088', '10000000-0000-0000-0000-000000000088', 'google', '100000000000000000088', 'zara.mensah@nexus.io',          now() - INTERVAL '2 days',  now() - INTERVAL '25 days'),
('20000000-0000-0000-0000-000000000089', '10000000-0000-0000-0000-000000000089', 'google', '100000000000000000089', 'adam.kowalski@nexus.io',        now() - INTERVAL '1 day',   now() - INTERVAL '20 days'),
('20000000-0000-0000-0000-000000000090', '10000000-0000-0000-0000-000000000090', 'google', '100000000000000000090', 'bella.nkosi@nexus.io',          now() - INTERVAL '3 days',  now() - INTERVAL '15 days'),
-- Inactive users: last login was long ago
('20000000-0000-0000-0000-000000000091', '10000000-0000-0000-0000-000000000091', 'google', '100000000000000000091', 'chris.olawale@nexus.io',        now() - INTERVAL '200 days', now() - INTERVAL '400 days'),
('20000000-0000-0000-0000-000000000092', '10000000-0000-0000-0000-000000000092', 'google', '100000000000000000092', 'diana.fischer@nexus.io',        now() - INTERVAL '180 days', now() - INTERVAL '380 days'),
('20000000-0000-0000-0000-000000000093', '10000000-0000-0000-0000-000000000093', 'google', '100000000000000000093', 'evan.oduola@nexus.io',          now() - INTERVAL '160 days', now() - INTERVAL '360 days'),
('20000000-0000-0000-0000-000000000094', '10000000-0000-0000-0000-000000000094', 'google', '100000000000000000094', 'faye.leclerc@nexus.io',         now() - INTERVAL '120 days', now() - INTERVAL '300 days'),
('20000000-0000-0000-0000-000000000095', '10000000-0000-0000-0000-000000000095', 'google', '100000000000000000095', 'glen.abubakar@nexus.io',        now() - INTERVAL '100 days', now() - INTERVAL '280 days'),
-- Contractors
('20000000-0000-0000-0000-000000000096', '10000000-0000-0000-0000-000000000096', 'google', '100000000000000000096', 'hana.schmidt@contractor.io',    now() - INTERVAL '1 day',   now() - INTERVAL '10 days'),
('20000000-0000-0000-0000-000000000097', '10000000-0000-0000-0000-000000000097', 'google', '100000000000000000097', 'ivan.petrov@contractor.io',     now() - INTERVAL '2 days',  now() - INTERVAL '8 days'),
('20000000-0000-0000-0000-000000000098', '10000000-0000-0000-0000-000000000098', 'google', '100000000000000000098', 'jade.owusu@contractor.io',      now() - INTERVAL '1 day',   now() - INTERVAL '6 days'),
('20000000-0000-0000-0000-000000000099', '10000000-0000-0000-0000-000000000099', 'google', '100000000000000000099', 'kai.bergstrom@contractor.io',   now() - INTERVAL '1 day',   now() - INTERVAL '4 days'),
('20000000-0000-0000-0000-000000000100', '10000000-0000-0000-0000-000000000100', 'google', '100000000000000000100', 'luna.diaz@contractor.io',       now() - INTERVAL '1 day',   now() - INTERVAL '2 days');

-- =============================================================================
-- 6. OAUTH_STATES  (a few active pending states + expired ones)
-- =============================================================================
INSERT INTO oauth_states (state, code_verifier, redirect_to, created_at, expires_at) VALUES
-- Active states (not yet consumed — browser mid-flow)
('a3f8c1e2d4b6f7a9c0e1d2b3f4a5c6e7', 'dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk', '/dashboard',       now() - INTERVAL '1 minute',  now() + INTERVAL '4 minutes'),
('b5d9e0f1a2c3e4d5b6f7a8c9d0e1f2a3', 'M25iVXpKU3puUjFaYjE2eXBoYzdVOGNLQ2F2dXdlZUk', '/reports',         now() - INTERVAL '2 minutes', now() + INTERVAL '3 minutes'),
('c7e1f2a3b4d5c6e7f8a9b0c1d2e3f4a5', 'c29tZV9yYW5kb21fdmVyaWZpZXJfZm9yX3BrY2Vfc2Vj', '/settings/profile', now() - INTERVAL '30 seconds', now() + INTERVAL '4 minutes 30 seconds'),
-- Expired states (should be cleaned up by maintenance job)
('d8f2a3b4c5e6d7f8a9b0c1d2e3f4a5b6', 'ZXhwaXJlZF92ZXJpZmllcl8xMjM0NTY3ODkwYWJjZA',  '/admin',           now() - INTERVAL '10 minutes', now() - INTERVAL '5 minutes'),
('e9a3b4c5d6f7e8a9b0c1d2e3f4a5b6c7', 'YW5vdGhlcl9leHBpcmVkX3ZlcmlmaWVyX3h5ejEyMw',  '/billing',         now() - INTERVAL '15 minutes', now() - INTERVAL '10 minutes');

-- =============================================================================
-- 7. SESSIONS
--    Mix of: active, expired, revoked, multi-device
-- =============================================================================
INSERT INTO sessions (id, user_id, ip_address, user_agent, is_revoked, absolute_expiry, created_at, last_seen_at) VALUES

-- ── Active sessions (currently logged in) ─────────────────────────────────────
('30000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '203.0.113.10',   'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/537.36 Chrome/120.0',        false, now() + INTERVAL '29 days', now() - INTERVAL '1 day',   now() - INTERVAL '2 hours'),
('30000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', '198.51.100.22',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/121.0',           false, now() + INTERVAL '28 days', now() - INTERVAL '2 days',  now() - INTERVAL '5 hours'),
('30000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000003', '192.0.2.45',     'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/120.0',                    false, now() + INTERVAL '27 days', now() - INTERVAL '3 days',  now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000004', '203.0.113.55',   'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_1) Safari/605.1.15',                       false, now() + INTERVAL '29 days', now() - INTERVAL '1 day',   now() - INTERVAL '30 minutes'),
('30000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000005', '198.51.100.88',  'Mozilla/5.0 (iPhone; CPU iPhone OS 17_2) AppleWebKit/605.1.15 Mobile Safari',         false, now() + INTERVAL '26 days', now() - INTERVAL '4 days',  now() - INTERVAL '3 hours'),
('30000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000006', '192.0.2.101',    'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0',   false, now() + INTERVAL '28 days', now() - INTERVAL '2 days',  now() - INTERVAL '45 minutes'),
('30000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000007', '203.0.113.77',   'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/537.36 Chrome/120.0',        false, now() + INTERVAL '25 days', now() - INTERVAL '5 days',  now() - INTERVAL '2 hours'),
('30000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000008', '198.51.100.33',  'Mozilla/5.0 (Windows NT 11.0; Win64; x64) AppleWebKit/537.36 Chrome/121.0',           false, now() + INTERVAL '29 days', now() - INTERVAL '1 day',   now() - INTERVAL '10 minutes'),
('30000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000009', '192.0.2.200',    'Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0',     false, now() + INTERVAL '24 days', now() - INTERVAL '6 days',  now() - INTERVAL '4 hours'),
('30000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000010', '203.0.113.120',  'Mozilla/5.0 (iPad; CPU OS 17_1) AppleWebKit/605.1.15 Mobile Safari',                  false, now() + INTERVAL '27 days', now() - INTERVAL '3 days',  now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000011', '198.51.100.150', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 13_6) Safari/605.1.15',                       false, now() + INTERVAL '28 days', now() - INTERVAL '2 days',  now() - INTERVAL '20 minutes'),
('30000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000012', '192.0.2.78',     'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Edge/120.0',            false, now() + INTERVAL '29 days', now() - INTERVAL '1 day',   now() - INTERVAL '15 minutes'),
('30000000-0000-0000-0000-000000000013', '10000000-0000-0000-0000-000000000013', '203.0.113.200',  'Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 Chrome/120.0 Mobile',              false, now() + INTERVAL '23 days', now() - INTERVAL '7 days',  now() - INTERVAL '6 hours'),
('30000000-0000-0000-0000-000000000014', '10000000-0000-0000-0000-000000000014', '198.51.100.50',  'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/537.36 Chrome/121.0',        false, now() + INTERVAL '27 days', now() - INTERVAL '3 days',  now() - INTERVAL '2 hours'),
('30000000-0000-0000-0000-000000000015', '10000000-0000-0000-0000-000000000015', '192.0.2.33',     'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0',   false, now() + INTERVAL '26 days', now() - INTERVAL '4 days',  now() - INTERVAL '3 hours'),

-- Multi-device: user 001 (admin) logged in from two devices simultaneously
('30000000-0000-0000-0000-000000000050', '10000000-0000-0000-0000-000000000001', '10.0.0.5',       'Mozilla/5.0 (iPhone; CPU iPhone OS 17_2) AppleWebKit/605.1.15 Mobile Safari',         false, now() + INTERVAL '28 days', now() - INTERVAL '2 days',  now() - INTERVAL '1 hour'),

-- More active sessions (users 16–90)
('30000000-0000-0000-0000-000000000016', '10000000-0000-0000-0000-000000000016', '203.0.113.16',   'Mozilla/5.0 (Macintosh) Chrome/120.0',  false, now() + INTERVAL '29 days', now() - INTERVAL '1 day',  now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000017', '10000000-0000-0000-0000-000000000017', '198.51.100.17',  'Mozilla/5.0 (Windows) Chrome/121.0',    false, now() + INTERVAL '28 days', now() - INTERVAL '2 days', now() - INTERVAL '2 hours'),
('30000000-0000-0000-0000-000000000018', '10000000-0000-0000-0000-000000000018', '192.0.2.18',     'Mozilla/5.0 (Linux) Firefox/122.0',     false, now() + INTERVAL '27 days', now() - INTERVAL '3 days', now() - INTERVAL '3 hours'),
('30000000-0000-0000-0000-000000000019', '10000000-0000-0000-0000-000000000019', '203.0.113.19',   'Mozilla/5.0 (Macintosh) Safari/605.1',  false, now() + INTERVAL '26 days', now() - INTERVAL '4 days', now() - INTERVAL '4 hours'),
('30000000-0000-0000-0000-000000000020', '10000000-0000-0000-0000-000000000020', '198.51.100.20',  'Mozilla/5.0 (Windows) Edge/120.0',      false, now() + INTERVAL '25 days', now() - INTERVAL '5 days', now() - INTERVAL '5 hours'),
('30000000-0000-0000-0000-000000000021', '10000000-0000-0000-0000-000000000021', '203.0.113.21',   'Mozilla/5.0 (Android) Chrome/120.0',   false, now() + INTERVAL '29 days', now() - INTERVAL '1 day',  now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000022', '10000000-0000-0000-0000-000000000022', '192.0.2.22',     'Mozilla/5.0 (iPhone) Safari/605.1',    false, now() + INTERVAL '28 days', now() - INTERVAL '2 days', now() - INTERVAL '30 minutes'),
('30000000-0000-0000-0000-000000000023', '10000000-0000-0000-0000-000000000023', '198.51.100.23',  'Mozilla/5.0 (Windows) Chrome/121.0',   false, now() + INTERVAL '27 days', now() - INTERVAL '3 days', now() - INTERVAL '2 hours'),
('30000000-0000-0000-0000-000000000024', '10000000-0000-0000-0000-000000000024', '203.0.113.24',   'Mozilla/5.0 (Macintosh) Chrome/122.0', false, now() + INTERVAL '26 days', now() - INTERVAL '4 days', now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000025', '10000000-0000-0000-0000-000000000025', '192.0.2.25',     'Mozilla/5.0 (Linux) Firefox/121.0',    false, now() + INTERVAL '25 days', now() - INTERVAL '5 days', now() - INTERVAL '4 hours'),
('30000000-0000-0000-0000-000000000026', '10000000-0000-0000-0000-000000000026', '198.51.100.26',  'Mozilla/5.0 (Windows) Edge/121.0',     false, now() + INTERVAL '29 days', now() - INTERVAL '1 day',  now() - INTERVAL '15 minutes'),
('30000000-0000-0000-0000-000000000027', '10000000-0000-0000-0000-000000000027', '203.0.113.27',   'Mozilla/5.0 (Macintosh) Safari/605.1', false, now() + INTERVAL '24 days', now() - INTERVAL '6 days', now() - INTERVAL '3 hours'),
('30000000-0000-0000-0000-000000000028', '10000000-0000-0000-0000-000000000028', '192.0.2.28',     'Mozilla/5.0 (Android) Chrome/120.0',   false, now() + INTERVAL '27 days', now() - INTERVAL '3 days', now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000029', '10000000-0000-0000-0000-000000000029', '198.51.100.29',  'Mozilla/5.0 (Windows) Firefox/122.0',  false, now() + INTERVAL '26 days', now() - INTERVAL '4 days', now() - INTERVAL '2 hours'),
('30000000-0000-0000-0000-000000000030', '10000000-0000-0000-0000-000000000030', '203.0.113.30',   'Mozilla/5.0 (Macintosh) Chrome/120.0', false, now() + INTERVAL '29 days', now() - INTERVAL '1 day',  now() - INTERVAL '45 minutes'),
('30000000-0000-0000-0000-000000000031', '10000000-0000-0000-0000-000000000031', '192.0.2.31',     'Mozilla/5.0 (Windows) Chrome/121.0',   false, now() + INTERVAL '28 days', now() - INTERVAL '2 days', now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000032', '10000000-0000-0000-0000-000000000032', '198.51.100.32',  'Mozilla/5.0 (Linux) Firefox/120.0',    false, now() + INTERVAL '27 days', now() - INTERVAL '3 days', now() - INTERVAL '2 hours'),
('30000000-0000-0000-0000-000000000033', '10000000-0000-0000-0000-000000000033', '203.0.113.33',   'Mozilla/5.0 (Macintosh) Safari/605.1', false, now() + INTERVAL '26 days', now() - INTERVAL '4 days', now() - INTERVAL '30 minutes'),
('30000000-0000-0000-0000-000000000034', '10000000-0000-0000-0000-000000000034', '192.0.2.34',     'Mozilla/5.0 (iPad) Safari/605.1',      false, now() + INTERVAL '29 days', now() - INTERVAL '1 day',  now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000035', '10000000-0000-0000-0000-000000000035', '198.51.100.35',  'Mozilla/5.0 (Windows) Edge/122.0',     false, now() + INTERVAL '25 days', now() - INTERVAL '5 days', now() - INTERVAL '3 hours'),
('30000000-0000-0000-0000-000000000036', '10000000-0000-0000-0000-000000000036', '203.0.113.36',   'Mozilla/5.0 (Macintosh) Chrome/121.0', false, now() + INTERVAL '28 days', now() - INTERVAL '2 days', now() - INTERVAL '2 hours'),
('30000000-0000-0000-0000-000000000037', '10000000-0000-0000-0000-000000000037', '192.0.2.37',     'Mozilla/5.0 (Linux) Chrome/120.0',     false, now() + INTERVAL '27 days', now() - INTERVAL '3 days', now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000038', '10000000-0000-0000-0000-000000000038', '198.51.100.38',  'Mozilla/5.0 (Windows) Firefox/121.0',  false, now() + INTERVAL '26 days', now() - INTERVAL '4 days', now() - INTERVAL '4 hours'),
('30000000-0000-0000-0000-000000000039', '10000000-0000-0000-0000-000000000039', '203.0.113.39',   'Mozilla/5.0 (Macintosh) Safari/605.1', false, now() + INTERVAL '29 days', now() - INTERVAL '1 day',  now() - INTERVAL '30 minutes'),
('30000000-0000-0000-0000-000000000040', '10000000-0000-0000-0000-000000000040', '192.0.2.40',     'Mozilla/5.0 (Android) Chrome/121.0',   false, now() + INTERVAL '28 days', now() - INTERVAL '2 days', now() - INTERVAL '2 hours'),

-- ── Revoked sessions (logged out or force-terminated) ─────────────────────────
('30000000-0000-0000-0000-000000000101', '10000000-0000-0000-0000-000000000001', '203.0.113.10',   'Mozilla/5.0 (Macintosh) Chrome/119.0', true,  now() + INTERVAL '15 days', now() - INTERVAL '15 days', now() - INTERVAL '5 days'),
('30000000-0000-0000-0000-000000000102', '10000000-0000-0000-0000-000000000005', '198.51.100.88',  'Mozilla/5.0 (Windows) Chrome/118.0',   true,  now() + INTERVAL '10 days', now() - INTERVAL '20 days', now() - INTERVAL '10 days'),
('30000000-0000-0000-0000-000000000103', '10000000-0000-0000-0000-000000000010', '192.0.2.200',    'Mozilla/5.0 (Linux) Firefox/119.0',    true,  now() + INTERVAL '5 days',  now() - INTERVAL '25 days', now() - INTERVAL '15 days'),
('30000000-0000-0000-0000-000000000104', '10000000-0000-0000-0000-000000000015', '203.0.113.55',   'Mozilla/5.0 (Android) Chrome/117.0',   true,  now() + INTERVAL '8 days',  now() - INTERVAL '22 days', now() - INTERVAL '12 days'),
('30000000-0000-0000-0000-000000000105', '10000000-0000-0000-0000-000000000020', '198.51.100.22',  'Mozilla/5.0 (iPhone) Safari/604.1',    true,  now() + INTERVAL '12 days', now() - INTERVAL '18 days', now() - INTERVAL '8 days'),
-- Force-revoked due to suspicious login from unknown IP
('30000000-0000-0000-0000-000000000106', '10000000-0000-0000-0000-000000000003', '45.33.32.156',   'Mozilla/5.0 (Windows) Chrome/115.0',   true,  now() + INTERVAL '20 days', now() - INTERVAL '10 days', now() - INTERVAL '10 days'),

-- ── Expired sessions (past absolute_expiry, not yet cleaned) ──────────────────
('30000000-0000-0000-0000-000000000201', '10000000-0000-0000-0000-000000000021', '203.0.113.21',   'Mozilla/5.0 (Macintosh) Chrome/110.0', false, now() - INTERVAL '5 days',  now() - INTERVAL '35 days', now() - INTERVAL '5 days'),
('30000000-0000-0000-0000-000000000202', '10000000-0000-0000-0000-000000000022', '198.51.100.22',  'Mozilla/5.0 (Windows) Edge/110.0',     false, now() - INTERVAL '2 days',  now() - INTERVAL '32 days', now() - INTERVAL '3 days'),
('30000000-0000-0000-0000-000000000203', '10000000-0000-0000-0000-000000000023', '192.0.2.45',     'Mozilla/5.0 (Linux) Firefox/109.0',    false, now() - INTERVAL '10 days', now() - INTERVAL '40 days', now() - INTERVAL '10 days'),
('30000000-0000-0000-0000-000000000204', '10000000-0000-0000-0000-000000000024', '203.0.113.88',   'Mozilla/5.0 (Android) Chrome/109.0',   false, now() - INTERVAL '3 days',  now() - INTERVAL '33 days', now() - INTERVAL '4 days'),

-- Sessions for contractors
('30000000-0000-0000-0000-000000000096', '10000000-0000-0000-0000-000000000096', '85.214.132.117', 'Mozilla/5.0 (Windows) Chrome/121.0',   false, now() + INTERVAL '20 days', now() - INTERVAL '10 days', now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000097', '10000000-0000-0000-0000-000000000097', '46.101.166.19',  'Mozilla/5.0 (Linux) Firefox/122.0',    false, now() + INTERVAL '22 days', now() - INTERVAL '8 days',  now() - INTERVAL '2 hours'),
('30000000-0000-0000-0000-000000000098', '10000000-0000-0000-0000-000000000098', '159.203.95.45',  'Mozilla/5.0 (Macintosh) Safari/605.1', false, now() + INTERVAL '24 days', now() - INTERVAL '6 days',  now() - INTERVAL '3 hours'),
('30000000-0000-0000-0000-000000000099', '10000000-0000-0000-0000-000000000099', '165.22.122.9',   'Mozilla/5.0 (iPad) Safari/605.1',      false, now() + INTERVAL '26 days', now() - INTERVAL '4 days',  now() - INTERVAL '1 hour'),
('30000000-0000-0000-0000-000000000100', '10000000-0000-0000-0000-000000000100', '134.122.56.78',  'Mozilla/5.0 (Android) Chrome/121.0',   false, now() + INTERVAL '28 days', now() - INTERVAL '2 days',  now() - INTERVAL '30 minutes');

-- =============================================================================
-- 8. USER_ROLES
--    Realistic assignments: global roles + some scoped to org UUIDs
--    Org UUIDs used as scope_id (pretend they exist in an organizations table)
-- =============================================================================

-- Org scope IDs (fictional but consistent)
-- org-engineering:  'a0000000-0000-0000-0000-000000000001'
-- org-marketing:    'a0000000-0000-0000-0000-000000000002'
-- org-finance:      'a0000000-0000-0000-0000-000000000003'
-- org-legal:        'a0000000-0000-0000-0000-000000000004'
-- org-support:      'a0000000-0000-0000-0000-000000000005'
-- org-analytics:    'a0000000-0000-0000-0000-000000000006'
-- org-sales:        'a0000000-0000-0000-0000-000000000007'
-- org-design:       'a0000000-0000-0000-0000-000000000008'
-- =============================================================================
-- INSERT INTO user_roles
-- Columns: user_id, role_id, granted_by, granted_at
-- scope_type and scope_id omitted (no scoping)
-- =============================================================================

INSERT INTO user_roles (user_id, role_id, granted_by, granted_at) VALUES

-- ── Global Admins ─────────────────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000001', '00000001-0000-0000-0000-000000000001', NULL,                                   now() - INTERVAL '730 days'),
('10000000-0000-0000-0000-000000000002', '00000001-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '700 days'),
('10000000-0000-0000-0000-000000000003', '00000001-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '680 days'),

-- ── Auditor (Legal & Security) ────────────────────────────────────────────────
('10000000-0000-0000-0000-000000000003', '00000001-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '680 days'),
('10000000-0000-0000-0000-000000000038', '00000001-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '250 days'),
('10000000-0000-0000-0000-000000000039', '00000001-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '240 days'),
('10000000-0000-0000-0000-000000000040', '00000001-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '230 days'),

-- ── Engineering: Manager + Editors ───────────────────────────────────────────
('10000000-0000-0000-0000-000000000004', '00000001-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '650 days'),
('10000000-0000-0000-0000-000000000005', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '620 days'),
('10000000-0000-0000-0000-000000000006', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '610 days'),
('10000000-0000-0000-0000-000000000007', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '590 days'),
('10000000-0000-0000-0000-000000000008', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '580 days'),
('10000000-0000-0000-0000-000000000009', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '560 days'),
('10000000-0000-0000-0000-000000000010', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '540 days'),
('10000000-0000-0000-0000-000000000011', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '520 days'),
('10000000-0000-0000-0000-000000000012', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '510 days'),
('10000000-0000-0000-0000-000000000013', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '500 days'),
('10000000-0000-0000-0000-000000000014', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '490 days'),
('10000000-0000-0000-0000-000000000015', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '480 days'),

-- ── Product: Manager + Analysts ───────────────────────────────────────────────
('10000000-0000-0000-0000-000000000016', '00000001-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '470 days'),
('10000000-0000-0000-0000-000000000017', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000016', now() - INTERVAL '460 days'),
('10000000-0000-0000-0000-000000000018', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000016', now() - INTERVAL '450 days'),
('10000000-0000-0000-0000-000000000019', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000016', now() - INTERVAL '440 days'),
('10000000-0000-0000-0000-000000000020', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000016', now() - INTERVAL '430 days'),

-- ── Marketing: Manager + Editors ─────────────────────────────────────────────
('10000000-0000-0000-0000-000000000021', '00000001-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '420 days'),
('10000000-0000-0000-0000-000000000022', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000021', now() - INTERVAL '410 days'),
('10000000-0000-0000-0000-000000000023', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000021', now() - INTERVAL '400 days'),
('10000000-0000-0000-0000-000000000024', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000021', now() - INTERVAL '390 days'),
('10000000-0000-0000-0000-000000000025', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000021', now() - INTERVAL '380 days'),
('10000000-0000-0000-0000-000000000026', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000021', now() - INTERVAL '370 days'),
('10000000-0000-0000-0000-000000000027', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000021', now() - INTERVAL '360 days'),
('10000000-0000-0000-0000-000000000028', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000021', now() - INTERVAL '350 days'),
('10000000-0000-0000-0000-000000000029', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000021', now() - INTERVAL '340 days'),
('10000000-0000-0000-0000-000000000030', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000021', now() - INTERVAL '330 days'),

-- ── Finance: Billing Admins + Analysts ────────────────────────────────────────
('10000000-0000-0000-0000-000000000031', '00000001-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '320 days'),
('10000000-0000-0000-0000-000000000032', '00000001-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000031', now() - INTERVAL '310 days'),
('10000000-0000-0000-0000-000000000033', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000031', now() - INTERVAL '300 days'),
('10000000-0000-0000-0000-000000000034', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000031', now() - INTERVAL '290 days'),
('10000000-0000-0000-0000-000000000035', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000031', now() - INTERVAL '280 days'),
('10000000-0000-0000-0000-000000000036', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000031', now() - INTERVAL '270 days'),
('10000000-0000-0000-0000-000000000037', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000031', now() - INTERVAL '260 days'),

-- ── Support: Manager + Support Agents ─────────────────────────────────────────
('10000000-0000-0000-0000-000000000041', '00000001-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '220 days'),
('10000000-0000-0000-0000-000000000042', '00000001-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000041', now() - INTERVAL '215 days'),
('10000000-0000-0000-0000-000000000043', '00000001-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000041', now() - INTERVAL '210 days'),
('10000000-0000-0000-0000-000000000044', '00000001-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000041', now() - INTERVAL '205 days'),
('10000000-0000-0000-0000-000000000045', '00000001-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000041', now() - INTERVAL '200 days'),
('10000000-0000-0000-0000-000000000046', '00000001-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000041', now() - INTERVAL '195 days'),
('10000000-0000-0000-0000-000000000047', '00000001-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000041', now() - INTERVAL '190 days'),
('10000000-0000-0000-0000-000000000048', '00000001-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000041', now() - INTERVAL '185 days'),
('10000000-0000-0000-0000-000000000049', '00000001-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000041', now() - INTERVAL '180 days'),
('10000000-0000-0000-0000-000000000050', '00000001-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000041', now() - INTERVAL '175 days'),

-- ── Analytics: Manager + Analysts ─────────────────────────────────────────────
('10000000-0000-0000-0000-000000000051', '00000001-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '170 days'),
('10000000-0000-0000-0000-000000000052', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000051', now() - INTERVAL '165 days'),
('10000000-0000-0000-0000-000000000053', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000051', now() - INTERVAL '160 days'),
('10000000-0000-0000-0000-000000000054', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000051', now() - INTERVAL '155 days'),
('10000000-0000-0000-0000-000000000055', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000051', now() - INTERVAL '150 days'),
('10000000-0000-0000-0000-000000000056', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000051', now() - INTERVAL '145 days'),
('10000000-0000-0000-0000-000000000057', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000051', now() - INTERVAL '140 days'),
('10000000-0000-0000-0000-000000000058', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000051', now() - INTERVAL '135 days'),
('10000000-0000-0000-0000-000000000059', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000051', now() - INTERVAL '130 days'),
('10000000-0000-0000-0000-000000000060', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000051', now() - INTERVAL '125 days'),

-- ── Sales: Manager + Viewers + Analyst ────────────────────────────────────────
('10000000-0000-0000-0000-000000000061', '00000001-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '120 days'),
('10000000-0000-0000-0000-000000000062', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000061', now() - INTERVAL '118 days'),
('10000000-0000-0000-0000-000000000063', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000061', now() - INTERVAL '115 days'),
('10000000-0000-0000-0000-000000000064', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000061', now() - INTERVAL '112 days'),
('10000000-0000-0000-0000-000000000065', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000061', now() - INTERVAL '110 days'),
('10000000-0000-0000-0000-000000000066', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000061', now() - INTERVAL '108 days'),
('10000000-0000-0000-0000-000000000067', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000061', now() - INTERVAL '105 days'),
('10000000-0000-0000-0000-000000000068', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000061', now() - INTERVAL '102 days'),
('10000000-0000-0000-0000-000000000069', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000061', now() - INTERVAL '100 days'),
('10000000-0000-0000-0000-000000000070', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000061', now() - INTERVAL '98 days'),

-- ── Design: Manager + Editors + Viewer ────────────────────────────────────────
('10000000-0000-0000-0000-000000000071', '00000001-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '95 days'),
('10000000-0000-0000-0000-000000000072', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000071', now() - INTERVAL '92 days'),
('10000000-0000-0000-0000-000000000073', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000071', now() - INTERVAL '90 days'),
('10000000-0000-0000-0000-000000000074', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000071', now() - INTERVAL '87 days'),
('10000000-0000-0000-0000-000000000075', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000071', now() - INTERVAL '85 days'),

-- ── HR: Manager + Editor + Viewers ────────────────────────────────────────────
('10000000-0000-0000-0000-000000000076', '00000001-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '82 days'),
('10000000-0000-0000-0000-000000000077', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000076', now() - INTERVAL '80 days'),
('10000000-0000-0000-0000-000000000078', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000076', now() - INTERVAL '78 days'),
('10000000-0000-0000-0000-000000000079', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000076', now() - INTERVAL '75 days'),
('10000000-0000-0000-0000-000000000080', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000076', now() - INTERVAL '72 days'),

-- ── Recent Hires: viewer by default ──────────────────────────────────────────
('10000000-0000-0000-0000-000000000081', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '60 days'),
('10000000-0000-0000-0000-000000000082', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '55 days'),
('10000000-0000-0000-0000-000000000083', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '50 days'),
('10000000-0000-0000-0000-000000000084', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '45 days'),
('10000000-0000-0000-0000-000000000085', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '40 days'),
('10000000-0000-0000-0000-000000000086', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '35 days'),
('10000000-0000-0000-0000-000000000087', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '30 days'),
('10000000-0000-0000-0000-000000000088', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '25 days'),
('10000000-0000-0000-0000-000000000089', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '20 days'),
('10000000-0000-0000-0000-000000000090', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', now() - INTERVAL '15 days'),

-- ── Contractors: editor/analyst/viewer ────────────────────────────────────────
('10000000-0000-0000-0000-000000000096', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000021', now() - INTERVAL '10 days'),
('10000000-0000-0000-0000-000000000097', '00000001-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000004', now() - INTERVAL '8 days'),
('10000000-0000-0000-0000-000000000098', '00000001-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000051', now() - INTERVAL '6 days'),
('10000000-0000-0000-0000-000000000099', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000061', now() - INTERVAL '4 days'),
('10000000-0000-0000-0000-000000000100', '00000001-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000071', now() - INTERVAL '2 days');

COMMIT;

-- =============================================================================
-- VERIFICATION QUERIES  (run these to validate the seed)
-- =============================================================================

-- Row counts per table
-- SELECT 'users'            AS tbl, COUNT(*) FROM users
-- UNION ALL
-- SELECT 'oauth_accounts',           COUNT(*) FROM oauth_accounts
-- UNION ALL
-- SELECT 'oauth_states',             COUNT(*) FROM oauth_states
-- UNION ALL
-- SELECT 'sessions',                 COUNT(*) FROM sessions
-- UNION ALL
-- SELECT 'roles',                    COUNT(*) FROM roles
-- UNION ALL
-- SELECT 'permissions',              COUNT(*) FROM permissions
-- UNION ALL
-- SELECT 'role_permissions',         COUNT(*) FROM role_permissions
-- UNION ALL
-- SELECT 'user_roles',               COUNT(*) FROM user_roles;

-- Who has admin access?
-- SELECT u.display_name, u.email, r.name AS role
-- FROM user_roles ur
-- JOIN users u ON u.id = ur.user_id
-- JOIN roles r ON r.id = ur.role_id
-- WHERE r.name = 'admin';

-- Check permission resolution for a specific user:
-- SELECT p.resource, p.action
-- FROM user_roles ur
-- JOIN role_permissions rp ON rp.role_id = ur.role_id
-- JOIN permissions p ON p.id = rp.permission_id
-- WHERE ur.user_id = '10000000-0000-0000-0000-000000000022'
-- ORDER BY p.resource, p.action;

-- Active sessions with user info:
-- SELECT u.display_name, s.ip_address, s.created_at, s.last_seen_at
-- FROM sessions s
-- JOIN users u ON u.id = s.user_id
-- WHERE s.is_revoked = false AND s.absolute_expiry > now()
-- ORDER BY s.last_seen_at DESC;

-- Users with multiple active sessions:
-- SELECT u.display_name, COUNT(*) AS session_count
-- FROM sessions s
-- JOIN users u ON u.id = s.user_id
-- WHERE s.is_revoked = false AND s.absolute_expiry > now()
-- GROUP BY u.id, u.display_name
-- HAVING COUNT(*) > 1;
