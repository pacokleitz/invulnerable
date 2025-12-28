import { FC, useState, useEffect, useCallback } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { api } from '../../lib/api/client';
import type { Vulnerability } from '../../lib/api/types';
import { SeverityBadge } from '../ui/SeverityBadge';
import { StatusBadge } from '../ui/StatusBadge';
import { formatDate, daysSince, calculateSLAStatus } from '../../lib/utils/formatters';

export const VulnerabilitiesList: FC = () => {
	const [vulnerabilities, setVulnerabilities] = useState<Vulnerability[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [searchParams, setSearchParams] = useSearchParams();

	// Filters from URL
	const severityFilter = searchParams.get('severity') || '';
	const statusFilter = searchParams.get('status') || '';
	const packageFilter = searchParams.get('package') || '';
	const cveFilter = searchParams.get('cve') || '';
	const showUnfixed = searchParams.get('show_unfixed') === 'true'; // Default to false

	const loadVulnerabilities = useCallback(async () => {
		setLoading(true);
		setError(null);

		try {
			const params: { limit: number; severity?: string; status?: string; package_name?: string; cve_id?: string; has_fix?: boolean } = { limit: 200 };
			if (severityFilter) params.severity = severityFilter;
			if (statusFilter) params.status = statusFilter;
			if (packageFilter) params.package_name = packageFilter;
			if (cveFilter) params.cve_id = cveFilter;
			// When showUnfixed is false, only show CVEs with fixes (has_fix = true)
			if (!showUnfixed) params.has_fix = true;

			const data = await api.vulnerabilities.list(params);
			setVulnerabilities(data);
		} catch (e) {
			setError(e instanceof Error ? e.message : 'Failed to load vulnerabilities');
		} finally {
			setLoading(false);
		}
	}, [severityFilter, statusFilter, packageFilter, cveFilter, showUnfixed]);

	useEffect(() => {
		document.title = 'Vulnerabilities - Invulnerable';
	}, []);

	useEffect(() => {
		loadVulnerabilities();
	}, [loadVulnerabilities]);

	const updateFilter = useCallback((key: string, value: string) => {
		setSearchParams(prev => {
			const newParams = new URLSearchParams(prev);
			if (value) {
				newParams.set(key, value);
			} else {
				newParams.delete(key);
			}
			return newParams;
		});
	}, [setSearchParams]);

	const handleClearFilters = useCallback(() => {
		setSearchParams({});
	}, [setSearchParams]);

	if (loading) {
		return (
			<div className="text-center py-12">
				<p className="text-gray-500">Loading vulnerabilities...</p>
			</div>
		);
	}

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<h1 className="text-3xl font-bold text-gray-900">Vulnerabilities</h1>
			</div>

			{/* Filters */}
			<div className="card">
				<h3 className="text-sm font-semibold text-gray-700 mb-3">Filters</h3>
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
					<div>
						<label htmlFor="severity" className="block text-sm font-medium text-gray-700">
							Severity
						</label>
						<select
							id="severity"
							value={severityFilter}
							onChange={(e) => updateFilter('severity', e.target.value)}
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						>
							<option value="">All</option>
							<option value="Critical">Critical</option>
							<option value="High">High</option>
							<option value="Medium">Medium</option>
							<option value="Low">Low</option>
						</select>
					</div>

					<div>
						<label htmlFor="status" className="block text-sm font-medium text-gray-700">
							Status
						</label>
						<select
							id="status"
							value={statusFilter}
							onChange={(e) => updateFilter('status', e.target.value)}
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						>
							<option value="">All</option>
							<option value="active">Active</option>
							<option value="fixed">Fixed</option>
							<option value="ignored">Ignored</option>
							<option value="accepted">Accepted</option>
						</select>
					</div>

					<div>
						<label htmlFor="packageFilter" className="block text-sm font-medium text-gray-700">
							Package Name
						</label>
						<input
							type="text"
							id="packageFilter"
							value={packageFilter}
							onChange={(e) => updateFilter('package', e.target.value)}
							placeholder="e.g., openssl"
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						/>
					</div>

					<div>
						<label htmlFor="cveFilter" className="block text-sm font-medium text-gray-700">
							CVE ID
						</label>
						<input
							type="text"
							id="cveFilter"
							value={cveFilter}
							onChange={(e) => updateFilter('cve', e.target.value)}
							placeholder="e.g., CVE-2023-1234"
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						/>
					</div>
				</div>

				<div className="mt-4 flex justify-between items-center">
					<div className="flex items-center space-x-4">
						<p className="text-sm text-gray-600">Showing {vulnerabilities.length} vulnerabilities</p>
						<label className="flex items-center space-x-2 text-sm">
							<input
								type="checkbox"
								checked={showUnfixed}
								onChange={(e) => updateFilter('show_unfixed', e.target.checked ? 'true' : 'false')}
								className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
							/>
							<span className="text-gray-700">Show unfixed CVEs</span>
						</label>
					</div>
					<button onClick={handleClearFilters} className="btn btn-secondary text-sm">
						Clear Filters
					</button>
				</div>
			</div>

			{error && (
				<div className="card bg-red-50">
					<p className="text-red-600">{error}</p>
				</div>
			)}

			{vulnerabilities.length === 0 ? (
				<div className="card text-center py-12">
					<p className="text-gray-500">No vulnerabilities found</p>
				</div>
			) : (
				<div className="card overflow-hidden">
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Image
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										CVE ID
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Package
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Version
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Severity
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Status
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										First Detected / Age
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										SLA Status
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Fix Version
									</th>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{vulnerabilities.map((vuln) => {
									// Calculate SLA status if we have the necessary data
									const slaStatus = vuln.sla_critical && vuln.sla_high && vuln.sla_medium && vuln.sla_low
										? calculateSLAStatus(
												vuln.first_detected_at,
												vuln.severity,
												{
													critical: vuln.sla_critical,
													high: vuln.sla_high,
													medium: vuln.sla_medium,
													low: vuln.sla_low,
												}
										  )
										: null;

									return (
										<tr key={`${vuln.id}-${vuln.image_id || 'unknown'}`} className="hover:bg-gray-50">
											<td className="px-6 py-4 text-sm text-gray-900">
												{vuln.image_name ? (
													<Link
														to={`/images/${vuln.image_id}`}
														className="text-blue-600 hover:underline font-medium"
													>
														{vuln.image_name}
													</Link>
												) : (
													<span className="text-gray-400">N/A</span>
												)}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-blue-600">
												<Link
													to={`/vulnerabilities/${vuln.cve_id}`}
													className="hover:underline"
												>
													{vuln.cve_id}
												</Link>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
												{vuln.package_name}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
												{vuln.package_version}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<SeverityBadge severity={vuln.severity} />
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<StatusBadge status={vuln.status} />
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
												<div>
													{formatDate(vuln.first_detected_at)}
												</div>
												<div className="text-xs text-gray-400">
													({daysSince(vuln.first_detected_at)} days ago)
												</div>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												{slaStatus ? (
													<div className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${slaStatus.bgColor} ${slaStatus.color}`}>
														{slaStatus.status === 'exceeded' && (
															<span>Exceeded by {Math.abs(slaStatus.daysRemaining)} days</span>
														)}
														{slaStatus.status === 'warning' && (
															<span>{slaStatus.daysRemaining} days remaining</span>
														)}
														{slaStatus.status === 'compliant' && (
															<span>{slaStatus.daysRemaining} days remaining</span>
														)}
													</div>
												) : (
													<span className="text-gray-400">N/A</span>
												)}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
												{vuln.fix_version || 'N/A'}
											</td>
										</tr>
									);
								})}
							</tbody>
						</table>
					</div>
				</div>
			)}
		</div>
	);
};
