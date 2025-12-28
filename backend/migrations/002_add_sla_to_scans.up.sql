-- Add SLA configuration columns to scans table
-- These store the Service Level Agreement remediation deadlines in days for each severity level

ALTER TABLE scans
ADD COLUMN sla_critical INTEGER NOT NULL DEFAULT 7,
ADD COLUMN sla_high INTEGER NOT NULL DEFAULT 30,
ADD COLUMN sla_medium INTEGER NOT NULL DEFAULT 90,
ADD COLUMN sla_low INTEGER NOT NULL DEFAULT 180;
