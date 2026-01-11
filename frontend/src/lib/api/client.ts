import type {
	DashboardMetrics,
	ImageWithStats,
	PaginatedResponse,
	ScanDiff,
	ScanWithDetails,
	User,
	Vulnerability,
	VulnerabilityHistory,
	VulnerabilityUpdate
} from './types';

const API_BASE = '/api/v1';

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
	const response = await fetch(`${API_BASE}${endpoint}`, {
		headers: {
			'Content-Type': 'application/json',
			...options?.headers
		},
		...options
	});

	// Handle authentication failures - redirect to OAuth login
	if (response.status === 401) {
		// OAuth2 is enabled but user is not authenticated
		// Redirect to OAuth2 Proxy login with return URL
		const returnUrl = encodeURIComponent(window.location.pathname + window.location.search);
		window.location.href = `/oauth2/start?rd=${returnUrl}`;
		// Throw to prevent further processing
		throw new Error('Authentication required - redirecting to login');
	}

	if (!response.ok) {
		const error = await response.text();
		throw new Error(`API Error: ${response.status} - ${error}`);
	}

	return response.json();
}

export const api = {
	// Scans
	scans: {
		list: (params?: {
			limit?: number;
			offset?: number;
			image_id?: number;
			image?: string;
			from_date?: string;
			to_date?: string;
			has_fix?: boolean;
		}) => {
			const searchParams = new URLSearchParams();
			if (params?.limit) searchParams.set('limit', params.limit.toString());
			if (params?.offset) searchParams.set('offset', params.offset.toString());
			if (params?.image_id) searchParams.set('image_id', params.image_id.toString());
			if (params?.image) searchParams.set('image', params.image);
			if (params?.from_date) searchParams.set('from_date', params.from_date);
			if (params?.to_date) searchParams.set('to_date', params.to_date);
			if (params?.has_fix !== undefined) searchParams.set('has_fix', params.has_fix.toString());

			const query = searchParams.toString();
			return fetchAPI<PaginatedResponse<ScanWithDetails>>(`/scans${query ? `?${query}` : ''}`);
		},

		get: (id: number, has_fix?: boolean) => {
			const searchParams = new URLSearchParams();
			if (has_fix !== undefined) searchParams.set('has_fix', has_fix.toString());
			const query = searchParams.toString();
			return fetchAPI<{ scan: ScanWithDetails; vulnerabilities: Vulnerability[] }>(`/scans/${id}${query ? `?${query}` : ''}`);
		},

		getSBOM: (id: number) => {
			return fetchAPI<any>(`/scans/${id}/sbom`);
		},

		getDiff: (id: number, previousScanId?: number) => {
			const params = previousScanId ? `?previous_scan_id=${previousScanId}` : '';
			return fetchAPI<ScanDiff>(`/scans/${id}/diff${params}`);
		}
	},

	// Vulnerabilities
	vulnerabilities: {
		list: (params?: {
			limit?: number;
			offset?: number;
			severity?: string;
			status?: string;
			image_name?: string;
			cve_id?: string;
			image_id?: number;
			has_fix?: boolean;
		}) => {
			const searchParams = new URLSearchParams();
			if (params?.limit) searchParams.set('limit', params.limit.toString());
			if (params?.offset) searchParams.set('offset', params.offset.toString());
			if (params?.severity) searchParams.set('severity', params.severity);
			if (params?.status) searchParams.set('status', params.status);
			if (params?.image_name) searchParams.set('image_name', params.image_name);
			if (params?.cve_id) searchParams.set('cve_id', params.cve_id);
			if (params?.image_id) searchParams.set('image_id', params.image_id.toString());
			if (params?.has_fix !== undefined) searchParams.set('has_fix', params.has_fix.toString());

			const query = searchParams.toString();
			return fetchAPI<PaginatedResponse<Vulnerability>>(`/vulnerabilities${query ? `?${query}` : ''}`);
		},

		getByCVE: (cve: string) => {
			return fetchAPI<Vulnerability[]>(`/vulnerabilities/${cve}`);
		},

		update: (id: number, update: VulnerabilityUpdate) => {
			return fetchAPI<Vulnerability>(`/vulnerabilities/${id}`, {
				method: 'PATCH',
				body: JSON.stringify(update)
			});
		},

		bulkUpdate: (vulnerabilityIds: number[], update: VulnerabilityUpdate) => {
			return fetchAPI<{ updated_count: number; status?: string }>(`/vulnerabilities/bulk`, {
				method: 'PATCH',
				body: JSON.stringify({
					vulnerability_ids: vulnerabilityIds,
					...update
				})
			});
		},

		getHistory: (id: number) => {
			return fetchAPI<VulnerabilityHistory[]>(`/vulnerabilities/${id}/history`);
		}
	},

	// Images
	images: {
		list: (params?: {
			limit?: number;
			offset?: number;
			registry?: string;
			repository?: string;
			tag?: string;
			has_fix?: boolean;
		}) => {
			const searchParams = new URLSearchParams();
			if (params?.limit) searchParams.set('limit', params.limit.toString());
			if (params?.offset) searchParams.set('offset', params.offset.toString());
			if (params?.registry) searchParams.set('registry', params.registry);
			if (params?.repository) searchParams.set('repository', params.repository);
			if (params?.tag) searchParams.set('tag', params.tag);
			if (params?.has_fix !== undefined) searchParams.set('has_fix', params.has_fix.toString());

			const query = searchParams.toString();
			return fetchAPI<PaginatedResponse<ImageWithStats>>(`/images${query ? `?${query}` : ''}`);
		},

		getHistory: (id: number, limit?: number, offset?: number, has_fix?: boolean) => {
			const searchParams = new URLSearchParams();
			if (limit) searchParams.set('limit', limit.toString());
			if (offset) searchParams.set('offset', offset.toString());
			if (has_fix !== undefined) searchParams.set('has_fix', has_fix.toString());

			const query = searchParams.toString();
			return fetchAPI<PaginatedResponse<ScanWithDetails>>(`/images/${id}/history${query ? `?${query}` : ''}`);
		}
	},

	// Metrics
	metrics: {
		getDashboard: (has_fix?: boolean) => {
			const searchParams = new URLSearchParams();
			if (has_fix !== undefined) searchParams.set('has_fix', has_fix.toString());
			const query = searchParams.toString();
			return fetchAPI<DashboardMetrics>(`/metrics${query ? `?${query}` : ''}`);
		}
	},

	// User
	user: {
		getMe: async (): Promise<User | null> => {
			const response = await fetch(`${API_BASE}/user/me`, {
				headers: {
					'Content-Type': 'application/json'
				}
			});

			// 204 No Content means OAuth2 Proxy is not deployed (OAuth disabled)
			if (response.status === 204) {
				return null;
			}

			// Handle authentication failures - redirect to OAuth login
			if (response.status === 401) {
				// OAuth2 is enabled but user is not authenticated
				// Redirect to OAuth2 Proxy login with return URL
				const returnUrl = encodeURIComponent(window.location.pathname + window.location.search);
				window.location.href = `/oauth2/start?rd=${returnUrl}`;
				// Throw to prevent further processing
				throw new Error('Authentication required - redirecting to login');
			}

			if (!response.ok) {
				const error = await response.text();
				throw new Error(`API Error: ${response.status} - ${error}`);
			}

			return response.json();
		}
	}
};
