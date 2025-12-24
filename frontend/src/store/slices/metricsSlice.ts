import { StateCreator } from 'zustand';
import { api } from '../../lib/api/client';
import type { AppStore, MetricsState } from '../types';

export const createMetricsSlice: StateCreator<AppStore, [], [], MetricsState> = (set) => ({
	data: null,
	loading: false,
	error: null,

	loadMetrics: async () => {
		set({ loading: true, error: null });
		try {
			const data = await api.metrics.getDashboard();
			set({ data, loading: false });
		} catch (error) {
			set({
				error: error instanceof Error ? error.message : 'Failed to load metrics',
				loading: false
			});
		}
	}
});
