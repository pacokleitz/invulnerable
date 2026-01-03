-- Migration 005: Add onlyFixed filters to webhook configurations
-- This allows filtering webhook notifications to only include vulnerabilities with fixes available

ALTER TABLE imagescan_webhook_configs
ADD COLUMN IF NOT EXISTS scan_only_fixed BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS status_change_only_fixed BOOLEAN DEFAULT true;

COMMENT ON COLUMN imagescan_webhook_configs.scan_only_fixed IS 'When true, only send scan completion notifications for vulnerabilities with fixes available';
COMMENT ON COLUMN imagescan_webhook_configs.status_change_only_fixed IS 'When true, only send status change notifications for vulnerabilities with fixes available';
