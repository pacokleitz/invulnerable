-- Rollback migration 004: Remove CVE status change webhook support

-- Drop webhook configurations table
DROP INDEX IF EXISTS idx_webhook_configs_lookup;
DROP TABLE IF EXISTS imagescan_webhook_configs;

-- Remove ImageScan tracking from vulnerabilities
DROP INDEX IF EXISTS idx_vulnerabilities_imagescan;
ALTER TABLE vulnerabilities
DROP COLUMN IF EXISTS imagescan_name,
DROP COLUMN IF EXISTS imagescan_namespace;
