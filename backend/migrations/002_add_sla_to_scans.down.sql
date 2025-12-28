-- Rollback: Remove SLA configuration columns from scans table

ALTER TABLE scans
DROP COLUMN sla_critical,
DROP COLUMN sla_high,
DROP COLUMN sla_medium,
DROP COLUMN sla_low;
