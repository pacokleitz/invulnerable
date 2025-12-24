import { StateCreator } from 'zustand';
import { api } from '../../lib/api/client';
import type { AppStore, VulnerabilitiesState } from '../types';

export const createVulnerabilitiesSlice: StateCreator<
	AppStore,
	[],
	[],
	VulnerabilitiesState
> = (set, get) => ({
	vulnerabilities: [],
	loading: false,
	error: null,

	loadVulnerabilities: async (params) => {
		set({ loading: true, error: null });
		try {
			const vulnerabilities = await api.vulnerabilities.list(params);
			set({ vulnerabilities, loading: false });
		} catch (error) {
			set({
				error: error instanceof Error ? error.message : 'Failed to load vulnerabilities',
				loading: false
			});
		}
	},

	updateVulnerability: async (id, update) => {
		try {
			const updated = await api.vulnerabilities.update(id, update);
			const vulnerabilities = get().vulnerabilities.map((v) => (v.id === id ? updated : v));
			set({ vulnerabilities });
		} catch (error) {
			set({
				error: error instanceof Error ? error.message : 'Failed to update vulnerability'
			});
		}
	}
});
