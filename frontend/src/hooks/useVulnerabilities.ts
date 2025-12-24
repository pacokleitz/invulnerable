import { useEffect } from 'react';
import { useStore } from '../store';

export const useVulnerabilities = (params?: {
	limit?: number;
	offset?: number;
	severity?: string;
	status?: string;
}) => {
	const { vulnerabilities, loading, error, loadVulnerabilities, updateVulnerability } = useStore(
		(state) => ({
			vulnerabilities: state.vulnerabilities,
			loading: state.loading,
			error: state.error,
			loadVulnerabilities: state.loadVulnerabilities,
			updateVulnerability: state.updateVulnerability
		})
	);

	useEffect(() => {
		loadVulnerabilities(params);
	}, [loadVulnerabilities, params?.limit, params?.offset, params?.severity, params?.status]);

	return {
		vulnerabilities,
		loading,
		error,
		reload: () => loadVulnerabilities(params),
		updateVulnerability
	};
};
