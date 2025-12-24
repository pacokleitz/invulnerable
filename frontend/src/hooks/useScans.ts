import { useEffect, useMemo } from 'react';
import { useStore } from '../store';

export const useScans = (params?: { limit?: number; offset?: number; image_id?: number }) => {
	const { scans, loading, error, loadScans } = useStore((state) => ({
		scans: state.scans,
		loading: state.loading,
		error: state.error,
		loadScans: state.loadScans
	}));

	const stableParams = useMemo(() => params, [params?.limit, params?.offset, params?.image_id]);

	useEffect(() => {
		loadScans(stableParams);
	}, [loadScans, stableParams]);

	return { scans, loading, error, reload: () => loadScans(stableParams) };
};

export const useScan = (id: number) => {
	const { currentScan, loading, error, loadScan, clearCurrentScan } = useStore((state) => ({
		currentScan: state.currentScan,
		loading: state.loading,
		error: state.error,
		loadScan: state.loadScan,
		clearCurrentScan: state.clearCurrentScan
	}));

	useEffect(() => {
		loadScan(id);
		return () => clearCurrentScan();
	}, [id, loadScan, clearCurrentScan]);

	return { currentScan, loading, error, reload: () => loadScan(id) };
};
