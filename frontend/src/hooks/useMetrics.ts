import { useEffect } from 'react';
import { useStore } from '../store';

export const useMetrics = () => {
	const { data, loading, error, loadMetrics } = useStore((state) => ({
		data: state.data,
		loading: state.loading,
		error: state.error,
		loadMetrics: state.loadMetrics
	}));

	useEffect(() => {
		loadMetrics();
	}, [loadMetrics]);

	return { data, loading, error, reload: loadMetrics };
};
