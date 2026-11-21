-- Audit trail shared by every table: an AFTER UPDATE OR DELETE trigger on each
-- table writes the old and new row here. Reads version through jsonb so tables
-- without a version column can use the same function.
CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE IF NOT EXISTS audit.audit_logs (
    id           bigserial PRIMARY KEY,
    namespace    varchar(32) NOT NULL,
    table_schema varchar NOT NULL,
    table_name   varchar NOT NULL,
    operation    varchar NOT NULL,
    record_id    text,
    updated_by   text,
    old_version  int,
    new_data     jsonb,
    old_data     jsonb,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION public.log_audit() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        INSERT INTO audit.audit_logs (
            namespace, table_schema, table_name, operation, record_id, updated_by, old_version, old_data, new_data
        ) VALUES (
            NEW.namespace,
            TG_TABLE_SCHEMA,
            TG_TABLE_NAME,
            CASE WHEN NEW.deleted_at IS NOT NULL THEN 'SOFT_DELETE' ELSE 'UPDATE' END,
            OLD.id,
            NEW.updated_by,
            (to_jsonb(OLD) ->> 'version')::int,
            to_jsonb(OLD),
            to_jsonb(NEW)
        );
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit.audit_logs (
            namespace, table_schema, table_name, operation, record_id, old_version, old_data
        ) VALUES (
            OLD.namespace,
            TG_TABLE_SCHEMA,
            TG_TABLE_NAME,
            'HARD_DELETE',
            OLD.id,
            (to_jsonb(OLD) ->> 'version')::int,
            to_jsonb(OLD)
        );
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE INDEX IF NOT EXISTS audit_logs_namespace_schema_table_record_idx
    ON audit.audit_logs (namespace, table_schema, table_name, record_id);
