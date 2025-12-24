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

export const useImageHistory = (id: number, limit?: number) => {
	const { currentImageHistory, loading, error, loadImageHistory, clearImageHistory } = useStore(
		(state) => ({
			currentImageHistory: state.currentImageHistory,
			loading: state.loading,
			error: state.error,
			loadImageHistory: state.loadImageHistory,
			clearImageHistory: state.clearImageHistory
		})
	);

	useEffect(() => {
		loadImageHistory(id, limit);
		return () => clearImageHistory();
	}, [id, limit, loadImageHistory, clearImageHistory]);

	return { currentImageHistory, loading, error, reload: () => loadImageHistory(id, limit) };
};
