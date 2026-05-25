ALTER TABLE ignore_rules
    ADD COLUMN IF NOT EXISTS application_id UUID REFERENCES applications(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS resource_ref TEXT NULL;

ALTER TABLE ignore_rules
    DROP CONSTRAINT IF EXISTS chk_ignore_rules_resource_scope;

ALTER TABLE ignore_rules
    ADD CONSTRAINT chk_ignore_rules_resource_scope
    CHECK (resource_ref IS NULL OR application_id IS NOT NULL);

ALTER TABLE ignore_rules
    DROP CONSTRAINT IF EXISTS ignore_rules_workspace_id_name_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_ignore_rules_workspace_name
    ON ignore_rules (workspace_id, name)
    WHERE application_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_ignore_rules_application_name
    ON ignore_rules (application_id, name)
    WHERE application_id IS NOT NULL;

DROP INDEX IF EXISTS idx_ignore_rules_workspace_active;

CREATE INDEX IF NOT EXISTS idx_ignore_rules_analysis_active
    ON ignore_rules (workspace_id, application_id, created_at, id)
    WHERE active = TRUE;

CREATE INDEX IF NOT EXISTS idx_ignore_rules_application
    ON ignore_rules (workspace_id, application_id, active, created_at, id);
