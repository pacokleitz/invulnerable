import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { createMetricsSlice } from './slices/metricsSlice';
import { createScansSlice } from './slices/scansSlice';
import { createVulnerabilitiesSlice } from './slices/vulnerabilitiesSlice';
import { createImagesSlice } from './slices/imagesSlice';
import { createUserSlice } from './slices/userSlice';
import type { AppStore } from './types';

export const useStore = create<AppStore>()(
	devtools(
		(...a) => ({
			...createMetricsSlice(...a),
			...createScansSlice(...a),
			...createVulnerabilitiesSlice(...a),
			...createImagesSlice(...a),
			...createUserSlice(...a)
		}),
		{ name: 'InvulnerableStore' }
	)
);
