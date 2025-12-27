import { StateCreator } from 'zustand';
import { api } from '../../lib/api/client';
import type { AppStore, UserState } from '../types';

export const createUserSlice: StateCreator<AppStore, [], [], UserState> = (set) => ({
	user: null,
	loading: false,
	error: null,

	loadUser: async () => {
		set({ loading: true, error: null });
		try {
			const user = await api.user.getMe();
			set({ user, loading: false });
		} catch (error) {
			set({
				error: error instanceof Error ? error.message : 'Failed to load user',
				loading: false
			});
		}
	}
});
