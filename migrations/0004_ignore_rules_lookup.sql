CREATE INDEX IF NOT EXISTS idx_ignore_rules_workspace_active
    ON ignore_rules (workspace_id, created_at, id)
    WHERE active = TRUE;
