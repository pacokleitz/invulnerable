import { writable } from 'svelte/store';
import type { DashboardMetrics } from '../api/types';
import { api } from '../api/client';

export const metrics = writable<DashboardMetrics | null>(null);
export const metricsLoading = writable(false);
export const metricsError = writable<string | null>(null);

export async function loadMetrics() {
	metricsLoading.set(true);
	metricsError.set(null);

	try {
		const data = await api.metrics.getDashboard();
		metrics.set(data);
	} catch (error) {
		metricsError.set(error instanceof Error ? error.message : 'Failed to load metrics');
	} finally {
		metricsLoading.set(false);
	}
}
