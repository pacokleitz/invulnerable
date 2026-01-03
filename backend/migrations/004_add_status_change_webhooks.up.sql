-- Migration 004: Add support for CVE status change webhook notifications
-- This migration adds:
-- 1. ImageScan tracking on vulnerabilities (to know which ImageScan discovered each CVE)
-- 2. Webhook configuration storage (synced from ImageScan CRDs by the controller)

-- Track which ImageScan first discovered each vulnerability
-- This enables notification routing (only the discovering ImageScan gets notified)
ALTER TABLE vulnerabilities
ADD COLUMN imagescan_namespace VARCHAR(253),
ADD COLUMN imagescan_name VARCHAR(253);

CREATE INDEX idx_vulnerabilities_imagescan ON vulnerabilities(imagescan_namespace, imagescan_name);

COMMENT ON COLUMN vulnerabilities.imagescan_namespace IS 'Kubernetes namespace of ImageScan that first discovered this vulnerability';
COMMENT ON COLUMN vulnerabilities.imagescan_name IS 'Name of ImageScan resource that first discovered this vulnerability';

-- Store webhook configurations (synced from ImageScan CRDs by controller)
CREATE TABLE imagescan_webhook_configs (
    id SERIAL PRIMARY KEY,
    namespace VARCHAR(253) NOT NULL,
    name VARCHAR(253) NOT NULL,

    -- Main webhook configuration
    webhook_url VARCHAR(2048) NOT NULL,
    webhook_format VARCHAR(50) DEFAULT 'slack',

    -- Scan completion webhook settings
    scan_min_severity VARCHAR(50) DEFAULT 'High',

    -- Status change webhook settings
    status_change_enabled BOOLEAN DEFAULT false,
    status_change_min_severity VARCHAR(50) DEFAULT 'High',
    status_change_transitions TEXT[], -- Array of transitions like ["active→fixed", "active→ignored"]
    status_change_include_notes BOOLEAN DEFAULT false,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Ensure one config per ImageScan
    UNIQUE(namespace, name)
);

CREATE INDEX idx_webhook_configs_lookup ON imagescan_webhook_configs(namespace, name);

COMMENT ON TABLE imagescan_webhook_configs IS 'Webhook configurations synced from ImageScan CRDs';
COMMENT ON COLUMN imagescan_webhook_configs.status_change_transitions IS 'Array of status transitions to notify about (e.g., ["active→fixed"]). Empty means all transitions.';
