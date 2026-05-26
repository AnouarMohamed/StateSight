CREATE UNIQUE INDEX IF NOT EXISTS idx_clusters_workspace_id
    ON clusters (workspace_id, id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_source_definitions_workspace_id
    ON source_definitions (workspace_id, id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_applications_workspace_id
    ON applications (workspace_id, id);

ALTER TABLE applications
    ADD CONSTRAINT fk_applications_workspace_cluster
    FOREIGN KEY (workspace_id, cluster_id)
    REFERENCES clusters (workspace_id, id)
    ON DELETE CASCADE;

ALTER TABLE applications
    ADD CONSTRAINT fk_applications_workspace_source
    FOREIGN KEY (workspace_id, source_definition_id)
    REFERENCES source_definitions (workspace_id, id)
    ON DELETE CASCADE;

ALTER TABLE ignore_rules
    ADD CONSTRAINT fk_ignore_rules_workspace_application
    FOREIGN KEY (workspace_id, application_id)
    REFERENCES applications (workspace_id, id)
    ON DELETE CASCADE;
