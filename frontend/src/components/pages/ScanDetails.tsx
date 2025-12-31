import { FC, useEffect, useCallback, useState, useMemo } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useScan } from '../../hooks/useScans';
import { SeverityBadge } from '../ui/SeverityBadge';
import { StatusBadge } from '../ui/StatusBadge';
import { PackageCategoryBadge } from '../ui/PackageCategoryBadge';
import { VulnerabilityHistory } from '../ui/VulnerabilityHistory';
import { SortableTableHeader, useSortState } from '../ui/SortableTableHeader';
import { formatDate, daysSince, calculateSLAStatus } from '../../lib/utils/formatters';
import type { Vulnerability } from '../../lib/api/types';

export const ScanDetails: FC = () => {
	const { id } = useParams<{ id: string }>();
	const navigate = useNavigate();
	const scanId = parseInt(id || '0', 10);
	const [showUnfixed, setShowUnfixed] = useState(false);
	const [slaFilter, setSlaFilter] = useState<'all' | 'exceeded' | 'warning'>('all');
	const [historyVulnId, setHistoryVulnId] = useState<number | null>(null);
	const { sortKey, sortDirection, handleSort } = useSortState('severity', 'desc');
	const { currentScan, loading, error } = useScan(scanId, showUnfixed ? undefined : true);

	useEffect(() => {
		document.title = `Scan ${scanId} - Invulnerable`;
	}, [scanId]);

	const viewDiff = useCallback(() => {
		navigate(`/scans/${scanId}/diff`);
	}, [navigate, scanId]);

	const viewSBOM = useCallback(() => {
		navigate(`/scans/${scanId}/sbom`);
	}, [navigate, scanId]);

	// Filter and sort vulnerabilities - must be called before conditional returns (Rules of Hooks)
	const filteredVulnerabilities = useMemo(() => {
		if (!currentScan) return [];

		const { scan, vulnerabilities } = currentScan;
		// Filter based on showUnfixed checkbox
		let filtered = showUnfixed
			? vulnerabilities
			: vulnerabilities.filter(vuln => vuln.fix_version !== null && vuln.fix_version !== undefined);

		// Apply SLA filter
		if (slaFilter !== 'all') {
			filtered = filtered.filter(vuln => {
				const slaStatus = calculateSLAStatus(
					vuln.first_detected_at,
					vuln.severity,
					{
						critical: scan.sla_critical,
						high: scan.sla_high,
						medium: scan.sla_medium,
						low: scan.sla_low,
					},
					vuln.status,
					vuln.remediation_date,
					vuln.updated_at
				);
				if (slaFilter === 'exceeded') {
					return slaStatus.status === 'exceeded';
				} else if (slaFilter === 'warning') {
					return slaStatus.status === 'warning' || slaStatus.status === 'exceeded';
				}
				return true;
			});
		}

		// Apply sorting
		if (sortKey && sortDirection) {
			filtered.sort((a, b) => {
				let aVal: any;
				let bVal: any;

				// Handle SLA status FIRST (special ordering for compliance - not a real property)
				if (sortKey === 'sla_status') {
					const getSLASortValue = (vuln: typeof a) => {
						const slaStatus = calculateSLAStatus(
							vuln.first_detected_at,
							vuln.severity,
							{
								critical: scan.sla_critical,
								high: scan.sla_high,
								medium: scan.sla_medium,
								low: scan.sla_low,
							},
							vuln.status,
							vuln.remediation_date,
							vuln.updated_at
						);

						// Priority 1: Exceeded SLA (most overdue first)
						// Higher value = higher priority in descending sort
						// Range: 2000+ (more overdue = higher value)
						if (slaStatus.status === 'exceeded') {
							return 2000 + Math.abs(slaStatus.daysRemaining);
						}

						// Priority 2: Warning/Compliant (fewest days remaining first)
						// Fewer days = higher value = higher priority in descending sort
						// Range: 1-1000 (subtract from 1000 so fewer days = higher value)
						if (slaStatus.status === 'warning' || slaStatus.status === 'compliant') {
							return 1000 - slaStatus.daysRemaining;
						}

						// Priority 3: Fixed/Ignored/Accepted (at the bottom)
						// Lowest value = lowest priority
						return 0;
					};

					aVal = getSLASortValue(a);
					bVal = getSLASortValue(b);
				} else {
					// For all other fields, get the property value
					aVal = a[sortKey as keyof Vulnerability];
					bVal = b[sortKey as keyof Vulnerability];

					// Handle null/undefined values
					if (aVal == null && bVal == null) return 0;
					if (aVal == null) return 1;
					if (bVal == null) return -1;

					// Handle dates
					if (sortKey === 'first_detected_at' || sortKey === 'remediation_date') {
						aVal = new Date(aVal).getTime();
						bVal = new Date(bVal).getTime();
					}

					// Handle severity (special ordering)
					if (sortKey === 'severity') {
						const severityOrder = { Critical: 4, High: 3, Medium: 2, Low: 1, Negligible: 0, Unknown: -1 };
						aVal = severityOrder[aVal as keyof typeof severityOrder] || -1;
						bVal = severityOrder[bVal as keyof typeof severityOrder] || -1;
					}
				}

				// Compare
				if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
				if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
				return 0;
			});
		}

		return filtered;
	}, [currentScan, showUnfixed, slaFilter, sortKey, sortDirection]);

	if (loading) {
		return (
			<div className="text-center py-12">
				<p className="text-gray-500">Loading scan...</p>
			</div>
		);
	}

	if (error) {
		return (
			<div className="card bg-red-50">
				<p className="text-red-600">{error}</p>
			</div>
		);
	}

	if (!currentScan) {
		return null;
	}

	const { scan, vulnerabilities } = currentScan;

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<h1 className="text-3xl font-bold text-gray-900">Scan #{scan.id}</h1>
				<div className="space-x-2">
					<button onClick={viewSBOM} className="btn btn-secondary">
						View SBOM
					</button>
					<button onClick={viewDiff} className="btn btn-primary">
						View Diff
					</button>
				</div>
			</div>

			{/* Scan Details */}
			<div className="card">
				<h2 className="text-xl font-bold text-gray-900 mb-4">Scan Details</h2>
				<dl className="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<dt className="text-sm font-medium text-gray-500">Image</dt>
						<dd className="mt-1 text-sm text-gray-900">{scan.image_name}</dd>
					</div>
					<div>
						<dt className="text-sm font-medium text-gray-500">Image Digest</dt>
						<dd className="mt-1 text-sm text-gray-900 font-mono break-all">
							{scan.image_digest || 'N/A'}
						</dd>
					</div>
					<div>
						<dt className="text-sm font-medium text-gray-500">Scan Date</dt>
						<dd className="mt-1 text-sm text-gray-900">{formatDate(scan.scan_date)}</dd>
					</div>
					<div>
						<dt className="text-sm font-medium text-gray-500">Grype Version</dt>
						<dd className="mt-1 text-sm text-gray-900">{scan.grype_version || 'N/A'}</dd>
					</div>
					<div>
						<dt className="text-sm font-medium text-gray-500">Total Vulnerabilities</dt>
						<dd className="mt-1 text-sm text-gray-900">{scan.vulnerability_count}</dd>
					</div>
				</dl>
			</div>

			{/* Severity Summary */}
			<div className="card">
				<h2 className="text-xl font-bold text-gray-900 mb-4">Severity Summary</h2>
				<div className="grid grid-cols-1 md:grid-cols-4 gap-4">
					<div className="flex items-center justify-between p-4 bg-red-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-red-800">Critical ({scan.sla_critical} days SLA)</p>
							<p className="text-2xl font-bold text-red-900">{scan.critical_count}</p>
						</div>
					</div>
					<div className="flex items-center justify-between p-4 bg-orange-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-orange-800">High ({scan.sla_high} days SLA)</p>
							<p className="text-2xl font-bold text-orange-900">{scan.high_count}</p>
						</div>
					</div>
					<div className="flex items-center justify-between p-4 bg-yellow-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-yellow-800">Medium ({scan.sla_medium} days SLA)</p>
							<p className="text-2xl font-bold text-yellow-900">{scan.medium_count}</p>
						</div>
					</div>
					<div className="flex items-center justify-between p-4 bg-blue-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-blue-800">Low ({scan.sla_low} days SLA)</p>
							<p className="text-2xl font-bold text-blue-900">{scan.low_count}</p>
						</div>
					</div>
				</div>
			</div>

			{/* Vulnerabilities List */}
			<div className="card">
				<div className="flex justify-between items-center mb-4">
					<h2 className="text-xl font-bold text-gray-900">Vulnerabilities</h2>
					<div className="flex items-center space-x-4">
						<label className="flex items-center space-x-2 text-sm">
							<input
								type="checkbox"
								checked={showUnfixed}
								onChange={(e) => setShowUnfixed(e.target.checked)}
								className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
							/>
							<span className="text-gray-700">Show unfixed CVEs</span>
						</label>
						<div className="flex items-center space-x-2 text-sm">
							<label className="text-gray-700">SLA Filter:</label>
							<select
								value={slaFilter}
								onChange={(e) => setSlaFilter(e.target.value as 'all' | 'exceeded' | 'warning')}
								className="rounded border-gray-300 text-sm focus:ring-blue-500 focus:border-blue-500"
							>
								<option value="all">All</option>
								<option value="warning">At Risk</option>
								<option value="exceeded">Exceeded Only</option>
							</select>
						</div>
					</div>
				</div>
				<p className="text-sm text-gray-600 mb-4">
					Showing {filteredVulnerabilities.length} of {vulnerabilities.length} vulnerabilities
				</p>
				{filteredVulnerabilities.length === 0 ? (
					<p className="text-gray-500">No vulnerabilities found</p>
				) : (
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<SortableTableHeader
										label="CVE ID"
										sortKey="cve_id"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
									<SortableTableHeader
										label="Package"
										sortKey="package_name"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
									<SortableTableHeader
										label="Version"
										sortKey="package_version"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
									<SortableTableHeader
										label="Type"
										sortKey="package_type"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
									<SortableTableHeader
										label="Severity"
										sortKey="severity"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
									<SortableTableHeader
										label="Status"
										sortKey="status"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
									<SortableTableHeader
										label="First Detected / Age"
										sortKey="first_detected_at"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
									<SortableTableHeader
										label="SLA Status"
										sortKey="sla_status"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
									<SortableTableHeader
										label="Fix Version"
										sortKey="fix_version"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{filteredVulnerabilities.map((vuln) => {
									const slaStatus = calculateSLAStatus(
										vuln.first_detected_at,
										vuln.severity,
										{
											critical: scan.sla_critical,
											high: scan.sla_high,
											medium: scan.sla_medium,
											low: scan.sla_low,
										},
										vuln.status,
										vuln.remediation_date,
										vuln.updated_at
									);

									return (
										<tr key={vuln.id} className="hover:bg-gray-50">
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
												<PackageCategoryBadge packageType={vuln.package_type} />
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<SeverityBadge severity={vuln.severity} />
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<StatusBadge
													status={vuln.status}
													onClick={() => setHistoryVulnId(vuln.id)}
												/>
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
												<div className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${slaStatus.bgColor} ${slaStatus.color}`}>
													{slaStatus.status === 'fixed' && slaStatus.daysToFix !== undefined && (
														<span>Fixed in {slaStatus.daysToFix} {slaStatus.daysToFix === 1 ? 'day' : 'days'}</span>
													)}
													{slaStatus.status === 'accepted' && slaStatus.daysToFix !== undefined && (
														<span>Accepted in {slaStatus.daysToFix} {slaStatus.daysToFix === 1 ? 'day' : 'days'}</span>
													)}
													{slaStatus.status === 'ignored' && slaStatus.daysToFix !== undefined && (
														<span>Ignored in {slaStatus.daysToFix} {slaStatus.daysToFix === 1 ? 'day' : 'days'}</span>
													)}
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
				)}
			</div>

			{/* Vulnerability History Modal */}
			{historyVulnId && (
				<VulnerabilityHistory
					vulnerabilityId={historyVulnId}
					onClose={() => setHistoryVulnId(null)}
				/>
			)}
		</div>
	);
};
