import { useEffect, useMemo } from 'react';
import { useStore } from '../store';

export const useImages = (params?: { limit?: number; offset?: number }) => {
	const { images, loading, error, loadImages } = useStore((state) => ({
		images: state.images,
		loading: state.loading,
		error: state.error,
		loadImages: state.loadImages
	}));

	const stableParams = useMemo(() => params, [params?.limit, params?.offset]);

	useEffect(() => {
		loadImages(stableParams);
	}, [loadImages, stableParams]);

	return { images, loading, error, reload: () => loadImages(stableParams) };
};

export const useImageHistory = (id: number, limit?: number, offset?: number, hasFix?: boolean) => {
	const { currentImageHistory, historyTotal, loading, error, loadImageHistory, clearImageHistory } = useStore(
		(state) => ({
			currentImageHistory: state.currentImageHistory,
			historyTotal: state.historyTotal,
			loading: state.loading,
			error: state.error,
			loadImageHistory: state.loadImageHistory,
			clearImageHistory: state.clearImageHistory
		})
	);

	useEffect(() => {
		loadImageHistory(id, limit, offset, hasFix);
		return () => clearImageHistory();
	}, [id, limit, offset, hasFix, loadImageHistory, clearImageHistory]);

	return { currentImageHistory, historyTotal, loading, error, reload: () => loadImageHistory(id, limit, offset, hasFix) };
};
