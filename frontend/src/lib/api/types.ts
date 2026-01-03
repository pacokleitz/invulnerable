export interface PaginatedResponse<T> {
	data: T[];
	total: number;
	limit: number;
	offset: number;
}

export interface Image {
	id: number;
	registry: string;
	repository: string;
	tag: string;
	digest?: string;
	created_at: string;
	updated_at: string;
}

export interface ImageWithStats extends Image {
	scan_count: number;
	last_scan_date?: string;
	critical_count: number;
	high_count: number;
	medium_count: number;
	low_count: number;
}

export interface Scan {
	id: number;
	image_id: number;
	scan_date: string;
	syft_version?: string;
	grype_version?: string;
	status: string;
	sla_critical: number;
	sla_high: number;
	sla_medium: number;
	sla_low: number;
	created_at: string;
	updated_at: string;
}

export interface ScanWithDetails extends Scan {
	image_name: string;
	image_digest?: string;
	vulnerability_count: number;
	critical_count: number;
	high_count: number;
	medium_count: number;
	low_count: number;
}

export interface Vulnerability {
	id: number;
	cve_id: string;
	package_name: string;
	package_version: string;
	package_type?: string;
	severity: string;
	fix_version?: string;
	url?: string;
	description?: string;
	status: string;
	first_detected_at: string;
	last_seen_at: string;
	remediation_date?: string;
	notes?: string;
	updated_by?: string;
	created_at: string;
	updated_at: string;
	// Image context for compliance tracking (when fetched with image info)
	image_id?: number;
	image_name?: string;
	image_digest?: string;
	first_detected_at_for_image?: string;
	latest_scan_id?: number;
	latest_scan_date?: string;
	sla_critical?: number;
	sla_high?: number;
	sla_medium?: number;
	sla_low?: number;
}

export interface ScanDiff {
	scan_id: number;
	previous_scan_id: number;
	new_vulnerabilities: Vulnerability[];
	fixed_vulnerabilities: Vulnerability[];
	persistent_vulnerabilities: Vulnerability[];
	summary: {
		new_count: number;
		fixed_count: number;
		persistent_count: number;
	};
}

export interface DashboardMetrics {
	total_images: number;
	total_scans: number;
	total_vulnerabilities: number;
	active_vulnerabilities: number;
	severity_counts: {
		critical: number;
		high: number;
		medium: number;
		low: number;
	};
	recent_scans_24h: number;
}

export interface VulnerabilityUpdate {
	status?: string;
	notes?: string;
}

export interface User {
	email: string;
	username?: string;
}

export interface VulnerabilityHistory {
	id: number;
	vulnerability_id: number;
	field_name: string;
	old_value?: string;
	new_value?: string;
	changed_by?: string;
	changed_at: string;
	image_id?: number;
	image_name?: string;
}
