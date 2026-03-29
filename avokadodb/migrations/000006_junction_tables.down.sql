SET lock_timeout = '1s';
SET statement_timeout = '5s';

-- squawk-ignore ban-drop-table
DROP TABLE IF EXISTS auth_group_permissions;
-- squawk-ignore ban-drop-table
DROP TABLE IF EXISTS auth_user_permissions;
-- squawk-ignore ban-drop-table
DROP TABLE IF EXISTS auth_user_groups;
