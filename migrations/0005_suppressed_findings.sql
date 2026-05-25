CREATE TABLE IF NOT EXISTS suppressed_findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    desired_snapshot_id UUID NOT NULL REFERENCES desired_snapshots(id) ON DELETE CASCADE,
    live_snapshot_id UUID NOT NULL REFERENCES live_snapshots(id) ON DELETE CASCADE,
    ignore_rule_id UUID REFERENCES ignore_rules(id) ON DELETE SET NULL,
    ignore_rule_name TEXT NOT NULL,
    ignore_rule_reason TEXT NOT NULL,
    title TEXT NOT NULL,
    category TEXT NOT NULL,
    severity TEXT NOT NULL,
    resource_ref TEXT NOT NULL,
    field_path TEXT NOT NULL,
    desired_value TEXT NOT NULL,
    live_value TEXT NOT NULL,
    difference_type TEXT NOT NULL,
    suppressed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_suppressed_findings_application
    ON suppressed_findings (application_id, suppressed_at DESC);
