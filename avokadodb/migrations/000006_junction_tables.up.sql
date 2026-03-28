SET lock_timeout = '1s';
SET statement_timeout = '5s';

CREATE TABLE IF NOT EXISTS auth_user_groups (
    auth_user_id BIGINT NOT NULL REFERENCES auth_user(id) ON DELETE CASCADE,
    auth_group_id BIGINT NOT NULL REFERENCES auth_group(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (auth_user_id, auth_group_id)
);

CREATE TABLE IF NOT EXISTS auth_user_permissions (
    auth_user_id BIGINT NOT NULL REFERENCES auth_user(id) ON DELETE CASCADE,
    auth_permission_id BIGINT NOT NULL REFERENCES auth_permission(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (auth_user_id, auth_permission_id)
);

CREATE TABLE IF NOT EXISTS auth_group_permissions (
    auth_group_id BIGINT NOT NULL REFERENCES auth_group(id) ON DELETE CASCADE,
    auth_permission_id BIGINT NOT NULL REFERENCES auth_permission(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (auth_group_id, auth_permission_id)
);

-- squawk-ignore require-concurrent-index-creation
CREATE INDEX IF NOT EXISTS idx_auth_user_groups_user_id ON auth_user_groups (auth_user_id);
-- squawk-ignore require-concurrent-index-creation
CREATE INDEX IF NOT EXISTS idx_auth_user_groups_group_id ON auth_user_groups (auth_group_id);
-- squawk-ignore require-concurrent-index-creation
CREATE INDEX IF NOT EXISTS idx_auth_user_permissions_user_id ON auth_user_permissions (auth_user_id);
-- squawk-ignore require-concurrent-index-creation
CREATE INDEX IF NOT EXISTS idx_auth_user_permissions_perm_id ON auth_user_permissions (auth_permission_id);
-- squawk-ignore require-concurrent-index-creation
CREATE INDEX IF NOT EXISTS idx_auth_group_permissions_group_id ON auth_group_permissions (auth_group_id);
-- squawk-ignore require-concurrent-index-creation
CREATE INDEX IF NOT EXISTS idx_auth_group_permissions_perm_id ON auth_group_permissions (auth_permission_id);
