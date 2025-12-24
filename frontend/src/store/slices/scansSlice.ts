import { StateCreator } from 'zustand';
import { api } from '../../lib/api/client';
import type { AppStore, ScansState } from '../types';

export const createScansSlice: StateCreator<AppStore, [], [], ScansState> = (set) => ({
	scans: [],
	currentScan: null,
	loading: false,
	error: null,

	loadScans: async (params) => {
		set({ loading: true, error: null });
		try {
			const scans = await api.scans.list(params);
			set({ scans, loading: false });
		} catch (error) {
			set({
				error: error instanceof Error ? error.message : 'Failed to load scans',
				loading: false
			});
		}
	},

	loadScan: async (id) => {
		set({ loading: true, error: null });
		try {
			const currentScan = await api.scans.get(id);
			set({ currentScan, loading: false });
		} catch (error) {
			set({
				error: error instanceof Error ? error.message : 'Failed to load scan',
				loading: false
			});
		}
	},

	clearCurrentScan: () => set({ currentScan: null })
});
