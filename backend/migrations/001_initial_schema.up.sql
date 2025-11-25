-- Images table
CREATE TABLE IF NOT EXISTS images (
    id SERIAL PRIMARY KEY,
    registry VARCHAR(255) NOT NULL,
    repository VARCHAR(255) NOT NULL,
    tag VARCHAR(128) NOT NULL,
    digest VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(registry, repository, tag)
);

CREATE INDEX idx_images_repository ON images(repository);
CREATE INDEX idx_images_tag ON images(tag);

-- Scans table
CREATE TABLE IF NOT EXISTS scans (
    id SERIAL PRIMARY KEY,
    image_id INTEGER NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    scan_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    syft_version VARCHAR(50),
    grype_version VARCHAR(50),
    status VARCHAR(50) DEFAULT 'completed',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_scans_image_id ON scans(image_id);
CREATE INDEX idx_scans_scan_date ON scans(scan_date DESC);

-- SBOMs table
CREATE TABLE IF NOT EXISTS sboms (
    id SERIAL PRIMARY KEY,
    scan_id INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    format VARCHAR(50) NOT NULL, -- 'cyclonedx' or 'spdx'
    version VARCHAR(20),
    document JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(scan_id)
);

CREATE INDEX idx_sboms_scan_id ON sboms(scan_id);
CREATE INDEX idx_sboms_document ON sboms USING GIN(document);

-- Vulnerabilities table
CREATE TABLE IF NOT EXISTS vulnerabilities (
    id SERIAL PRIMARY KEY,
    cve_id VARCHAR(50) NOT NULL,
    package_name VARCHAR(255) NOT NULL,
    package_version VARCHAR(128) NOT NULL,
    package_type VARCHAR(50),
    severity VARCHAR(20),
    fix_version VARCHAR(128),
    url TEXT,
    description TEXT,
    status VARCHAR(50) DEFAULT 'active', -- active, fixed, ignored, accepted
    first_detected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    remediation_date TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(cve_id, package_name, package_version)
);

CREATE INDEX idx_vulnerabilities_cve_id ON vulnerabilities(cve_id);
CREATE INDEX idx_vulnerabilities_severity ON vulnerabilities(severity);
CREATE INDEX idx_vulnerabilities_status ON vulnerabilities(status);
CREATE INDEX idx_vulnerabilities_package_name ON vulnerabilities(package_name);
CREATE INDEX idx_vulnerabilities_first_detected ON vulnerabilities(first_detected_at);

-- Scan vulnerabilities junction table
CREATE TABLE IF NOT EXISTS scan_vulnerabilities (
    id SERIAL PRIMARY KEY,
    scan_id INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    vulnerability_id INTEGER NOT NULL REFERENCES vulnerabilities(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(scan_id, vulnerability_id)
);

CREATE INDEX idx_scan_vulnerabilities_scan_id ON scan_vulnerabilities(scan_id);
CREATE INDEX idx_scan_vulnerabilities_vulnerability_id ON scan_vulnerabilities(vulnerability_id);
