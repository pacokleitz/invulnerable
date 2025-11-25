import type {
	DashboardMetrics,
	ImageWithStats,
	Scan,
	ScanDiff,
	ScanWithDetails,
	Vulnerability,
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

	if (!response.ok) {
		const error = await response.text();
		throw new Error(`API Error: ${response.status} - ${error}`);
	}

	return response.json();
}

export const api = {
	// Scans
	scans: {
		list: (params?: { limit?: number; offset?: number; image_id?: number }) => {
			const searchParams = new URLSearchParams();
			if (params?.limit) searchParams.set('limit', params.limit.toString());
			if (params?.offset) searchParams.set('offset', params.offset.toString());
			if (params?.image_id) searchParams.set('image_id', params.image_id.toString());

			const query = searchParams.toString();
			return fetchAPI<ScanWithDetails[]>(`/scans${query ? `?${query}` : ''}`);
		},

		get: (id: number) => {
			return fetchAPI<{ scan: ScanWithDetails; vulnerabilities: Vulnerability[] }>(`/scans/${id}`);
		},

		getSBOM: (id: number) => {
			return fetchAPI<any>(`/scans/${id}/sbom`);
		},

		getDiff: (id: number) => {
			return fetchAPI<ScanDiff>(`/scans/${id}/diff`);
		}
	},

	// Vulnerabilities
	vulnerabilities: {
		list: (params?: { limit?: number; offset?: number; severity?: string; status?: string }) => {
			const searchParams = new URLSearchParams();
			if (params?.limit) searchParams.set('limit', params.limit.toString());
			if (params?.offset) searchParams.set('offset', params.offset.toString());
			if (params?.severity) searchParams.set('severity', params.severity);
			if (params?.status) searchParams.set('status', params.status);

			const query = searchParams.toString();
			return fetchAPI<Vulnerability[]>(`/vulnerabilities${query ? `?${query}` : ''}`);
		},

		getByCVE: (cve: string) => {
			return fetchAPI<Vulnerability[]>(`/vulnerabilities/${cve}`);
		},

		update: (id: number, update: VulnerabilityUpdate) => {
			return fetchAPI<Vulnerability>(`/vulnerabilities/${id}`, {
				method: 'PATCH',
				body: JSON.stringify(update)
			});
		}
	},

	// Images
	images: {
		list: (params?: { limit?: number; offset?: number }) => {
			const searchParams = new URLSearchParams();
			if (params?.limit) searchParams.set('limit', params.limit.toString());
			if (params?.offset) searchParams.set('offset', params.offset.toString());

			const query = searchParams.toString();
			return fetchAPI<ImageWithStats[]>(`/images${query ? `?${query}` : ''}`);
		},

		getHistory: (id: number, limit?: number) => {
			const searchParams = new URLSearchParams();
			if (limit) searchParams.set('limit', limit.toString());

			const query = searchParams.toString();
			return fetchAPI<ScanWithDetails[]>(`/images/${id}/history${query ? `?${query}` : ''}`);
		}
	},

	// Metrics
	metrics: {
		getDashboard: () => {
			return fetchAPI<DashboardMetrics>('/metrics');
		}
	}
};
