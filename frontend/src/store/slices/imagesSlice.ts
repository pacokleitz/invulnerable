import { StateCreator } from 'zustand';
import { api } from '../../lib/api/client';
import type { AppStore, ImagesState } from '../types';

export const createImagesSlice: StateCreator<AppStore, [], [], ImagesState> = (set) => ({
	images: [],
	total: 0,
	currentImageHistory: [],
	historyTotal: 0,
	loading: false,
	error: null,

	loadImages: async (params) => {
		set({ loading: true, error: null });
		try {
			const response = await api.images.list(params);
			set({ images: response.data, total: response.total, loading: false });
		} catch (error) {
			set({
				error: error instanceof Error ? error.message : 'Failed to load images',
				loading: false
			});
		}
	},

	loadImageHistory: async (id: number, limit?: number, offset?: number, hasFix?: boolean) => {
		set({ loading: true, error: null });
		try {
			const response = await api.images.getHistory(id, limit, offset, hasFix);
			set({ currentImageHistory: response.data, historyTotal: response.total, loading: false });
		} catch (error) {
			set({
				error: error instanceof Error ? error.message : 'Failed to load image history',
				loading: false
			});
		}
	},

	clearImageHistory: () => set({ currentImageHistory: [], historyTotal: 0 })
});
