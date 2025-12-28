import { StateCreator } from 'zustand';
import { api } from '../../lib/api/client';
import type { AppStore, ImagesState } from '../types';

export const createImagesSlice: StateCreator<AppStore, [], [], ImagesState> = (set) => ({
	images: [],
	currentImageHistory: [],
	loading: false,
	error: null,

	loadImages: async (params) => {
		set({ loading: true, error: null });
		try {
			const images = await api.images.list(params);
			set({ images, loading: false });
		} catch (error) {
			set({
				error: error instanceof Error ? error.message : 'Failed to load images',
				loading: false
			});
		}
	},

	loadImageHistory: async (id: number, limit?: number, hasFix?: boolean) => {
		set({ loading: true, error: null });
		try {
			const currentImageHistory = await api.images.getHistory(id, limit, hasFix);
			set({ currentImageHistory, loading: false });
		} catch (error) {
			set({
				error: error instanceof Error ? error.message : 'Failed to load image history',
				loading: false
			});
		}
	},

	clearImageHistory: () => set({ currentImageHistory: [] })
});
