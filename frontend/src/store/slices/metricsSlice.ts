import { StateCreator } from 'zustand';
import { api } from '../../lib/api/client';
import type { AppStore, MetricsState } from '../types';

export const createMetricsSlice: StateCreator<AppStore, [], [], MetricsState> = (set) => ({
	data: null,
	loading: false,
	error: null,

	loadMetrics: async (hasFix?: boolean, imageName?: string) => {
		set({ loading: true, error: null });
		try {
			const data = await api.metrics.getDashboard(hasFix, imageName);
			set({ data, loading: false });
		} catch (error) {
			// Don't update state if we're being redirected to login
			// Keep loading state so user sees "Loading..." during redirect
			if (error instanceof Error && error.message.includes('Authentication required')) {
				// Redirect is happening in fetchAPI, keep loading state
				return;
			}
			set({
				error: error instanceof Error ? error.message : 'Failed to load metrics',
				loading: false
			});
		}
	}
});
