import type {
	DashboardMetrics,
	ScanWithDetails,
	User,
	Vulnerability,
	ImageWithStats
} from '../lib/api/types';

export interface MetricsState {
	data: DashboardMetrics | null;
	loading: boolean;
	error: string | null;
	loadMetrics: () => Promise<void>;
}

export interface ScansState {
	scans: ScanWithDetails[];
	currentScan: { scan: ScanWithDetails; vulnerabilities: Vulnerability[] } | null;
	loading: boolean;
	error: string | null;
	loadScans: (params?: { limit?: number; offset?: number; image_id?: number }) => Promise<void>;
	loadScan: (id: number) => Promise<void>;
	clearCurrentScan: () => void;
}

export interface VulnerabilitiesState {
	vulnerabilities: Vulnerability[];
	loading: boolean;
	error: string | null;
	loadVulnerabilities: (params?: {
		limit?: number;
		offset?: number;
		severity?: string;
		status?: string;
	}) => Promise<void>;
	updateVulnerability: (id: number, update: { status?: string; notes?: string }) => Promise<void>;
}

export interface ImagesState {
	images: ImageWithStats[];
	currentImageHistory: ScanWithDetails[];
	loading: boolean;
	error: string | null;
	loadImages: (params?: { limit?: number; offset?: number }) => Promise<void>;
	loadImageHistory: (id: number, limit?: number) => Promise<void>;
	clearImageHistory: () => void;
}

export interface UserState {
	user: User | null;
	loading: boolean;
	error: string | null;
	loadUser: () => Promise<void>;
}

export interface AppStore extends MetricsState, ScansState, VulnerabilitiesState, ImagesState, UserState {}
