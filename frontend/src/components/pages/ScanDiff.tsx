import { FC, useState, useEffect, useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../../lib/api/client';
import type { ScanDiff as ScanDiffType, Scan, Vulnerability } from '../../lib/api/types';
import { SeverityBadge } from '../ui/SeverityBadge';
import { StatusBadge } from '../ui/StatusBadge';
import { PackageCategoryBadge } from '../ui/PackageCategoryBadge';
import { VulnerabilityHistory } from '../ui/VulnerabilityHistory';
import { SortableTableHeader, useSortState } from '../ui/SortableTableHeader';
import { formatDate, daysSince, calculateSLAStatus } from '../../lib/utils/formatters';

export const ScanDiff: FC = () => {
	const { id } = useParams<{ id: string }>();
	const scanId = parseInt(id || '0', 10);
	const [diff, setDiff] = useState<ScanDiffType | null>(null);
	const [scan, setScan] = useState<Scan | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [showUnfixed, setShowUnfixed] = useState(false);
	const [historyVulnId, setHistoryVulnId] = useState<number | null>(null);

	// Separate sort states for each table
	const newVulnsSort = useSortState('severity', 'desc');
	const fixedVulnsSort = useSortState('severity', 'desc');
	const persistentVulnsSort = useSortState('sla_status', 'desc');

	useEffect(() => {
		document.title = `Scan Diff - Scan ${scanId} - Invulnerable`;

		const fetchData = async () => {
			try {
				const [diffData, scanResult] = await Promise.all([
					api.scans.getDiff(scanId),
					api.scans.get(scanId)
				]);
				setDiff(diffData);
				setScan(scanResult.scan);
			} catch (e) {
				setError(e instanceof Error ? e.message : 'Failed to load scan diff');
			} finally {
				setLoading(false);
			}
		};

		fetchData();
	}, [scanId]);

	// Helper function to sort vulnerabilities
	const sortVulnerabilities = (vulns: Vulnerability[], sortKey: string | null, sortDirection: 'asc' | 'desc' | null, slaConfig: any) => {
		if (!sortKey || !sortDirection) return vulns;

		return [...vulns].sort((a, b) => {
			let aVal: any;
			let bVal: any;

			// Handle SLA status FIRST (not a real property)
			if (sortKey === 'sla_status') {
				const getSLASortValue = (vuln: Vulnerability) => {
					const slaStatus = calculateSLAStatus(
						vuln.first_detected_at,
						vuln.severity,
						slaConfig,
						vuln.status,
						vuln.remediation_date,
						vuln.updated_at
					);

					if (slaStatus.status === 'exceeded') {
						return 2000 + Math.abs(slaStatus.daysRemaining);
					}
					if (slaStatus.status === 'warning' || slaStatus.status === 'compliant') {
						return 1000 - slaStatus.daysRemaining;
					}
					return 0;
				};

				aVal = getSLASortValue(a);
				bVal = getSLASortValue(b);
			} else {
				aVal = a[sortKey as keyof Vulnerability];
				bVal = b[sortKey as keyof Vulnerability];

				if (aVal == null && bVal == null) return 0;
				if (aVal == null) return 1;
				if (bVal == null) return -1;

				if (sortKey === 'first_detected_at' || sortKey === 'remediation_date') {
					aVal = new Date(aVal).getTime();
					bVal = new Date(bVal).getTime();
				}

				if (sortKey === 'severity') {
					const severityOrder = { Critical: 4, High: 3, Medium: 2, Low: 1, Negligible: 0, Unknown: -1 };
					aVal = severityOrder[aVal as keyof typeof severityOrder] || -1;
					bVal = severityOrder[bVal as keyof typeof severityOrder] || -1;
				}
			}

			if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
			if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
			return 0;
		});
	};

	// Filter and sort vulnerabilities
	const filteredNewVulns = useMemo(() => {
		const filtered = showUnfixed
			? diff?.new_vulnerabilities || []
			: (diff?.new_vulnerabilities || []).filter(v => v.fix_version !== null && v.fix_version !== undefined);

		return scan ? sortVulnerabilities(filtered, newVulnsSort.sortKey, newVulnsSort.sortDirection, {
			critical: scan.sla_critical,
			high: scan.sla_high,
			medium: scan.sla_medium,
			low: scan.sla_low,
		}) : filtered;
	}, [diff, showUnfixed, newVulnsSort.sortKey, newVulnsSort.sortDirection, scan]);

	const filteredFixedVulns = useMemo(() => {
		const filtered = showUnfixed
			? diff?.fixed_vulnerabilities || []
			: (diff?.fixed_vulnerabilities || []).filter(v => v.fix_version !== null && v.fix_version !== undefined);

		return scan ? sortVulnerabilities(filtered, fixedVulnsSort.sortKey, fixedVulnsSort.sortDirection, {
			critical: scan.sla_critical,
			high: scan.sla_high,
			medium: scan.sla_medium,
			low: scan.sla_low,
		}) : filtered;
	}, [diff, showUnfixed, fixedVulnsSort.sortKey, fixedVulnsSort.sortDirection, scan]);

	const filteredPersistentVulns = useMemo(() => {
		const filtered = showUnfixed
			? diff?.persistent_vulnerabilities || []
			: (diff?.persistent_vulnerabilities || []).filter(v => v.fix_version !== null && v.fix_version !== undefined);

		return scan ? sortVulnerabilities(filtered, persistentVulnsSort.sortKey, persistentVulnsSort.sortDirection, {
			critical: scan.sla_critical,
			high: scan.sla_high,
			medium: scan.sla_medium,
			low: scan.sla_low,
		}) : filtered;
	}, [diff, showUnfixed, persistentVulnsSort.sortKey, persistentVulnsSort.sortDirection, scan]);

	if (loading) {
		return (
			<div className="text-center py-12">
				<p className="text-gray-500">Loading scan comparison...</p>
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

	if (!diff || !scan) {
		return null;
	}

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<h1 className="text-3xl font-bold text-gray-900">Scan Comparison</h1>
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
					<Link to={`/scans/${scanId}`} className="text-blue-600 hover:text-blue-800">
						← Back to Scan
					</Link>
				</div>
			</div>

			{/* Summary Cards */}
			<div className="grid grid-cols-1 md:grid-cols-3 gap-6">
				<div className="card bg-green-50">
					<h3 className="text-sm font-medium text-green-800">Fixed Vulnerabilities</h3>
					<p className="mt-2 text-3xl font-bold text-green-900">{diff.summary.fixed_count}</p>
					<p className="mt-1 text-xs text-green-700">Resolved since previous scan</p>
				</div>

				<div className="card bg-red-50">
					<h3 className="text-sm font-medium text-red-800">New Vulnerabilities</h3>
					<p className="mt-2 text-3xl font-bold text-red-900">{diff.summary.new_count}</p>
					<p className="mt-1 text-xs text-red-700">Introduced since previous scan</p>
				</div>

				<div className="card bg-gray-50">
					<h3 className="text-sm font-medium text-gray-800">Persistent Vulnerabilities</h3>
					<p className="mt-2 text-3xl font-bold text-gray-900">{diff.summary.persistent_count}</p>
					<p className="mt-1 text-xs text-gray-700">Still present from previous scan</p>
				</div>
			</div>

			<div className="card">
				<p className="text-sm text-gray-600">
					Comparing Scan #{diff.scan_id} with Scan #{diff.previous_scan_id}
				</p>
			</div>

			{/* New Vulnerabilities */}
			{filteredNewVulns.length > 0 && (
				<div className="card">
					<h2 className="text-xl font-bold text-red-900 mb-4">
						New Vulnerabilities ({filteredNewVulns.length})
					</h2>
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<SortableTableHeader
										label="CVE ID"
										sortKey="cve_id"
										currentSortKey={newVulnsSort.sortKey}
										currentSortDirection={newVulnsSort.sortDirection}
										onSort={newVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Package"
										sortKey="package_name"
										currentSortKey={newVulnsSort.sortKey}
										currentSortDirection={newVulnsSort.sortDirection}
										onSort={newVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Version"
										sortKey="package_version"
										currentSortKey={newVulnsSort.sortKey}
										currentSortDirection={newVulnsSort.sortDirection}
										onSort={newVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Type"
										sortKey="package_type"
										currentSortKey={newVulnsSort.sortKey}
										currentSortDirection={newVulnsSort.sortDirection}
										onSort={newVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Severity"
										sortKey="severity"
										currentSortKey={newVulnsSort.sortKey}
										currentSortDirection={newVulnsSort.sortDirection}
										onSort={newVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Status"
										sortKey="status"
										currentSortKey={newVulnsSort.sortKey}
										currentSortDirection={newVulnsSort.sortDirection}
										onSort={newVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="First Detected / Age"
										sortKey="first_detected_at"
										currentSortKey={newVulnsSort.sortKey}
										currentSortDirection={newVulnsSort.sortDirection}
										onSort={newVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="SLA Status"
										sortKey="sla_status"
										currentSortKey={newVulnsSort.sortKey}
										currentSortDirection={newVulnsSort.sortDirection}
										onSort={newVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Fix Version"
										sortKey="fix_version"
										currentSortKey={newVulnsSort.sortKey}
										currentSortDirection={newVulnsSort.sortDirection}
										onSort={newVulnsSort.handleSort}
									/>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{filteredNewVulns.map((vuln) => {
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
												<Link to={`/vulnerabilities/${vuln.cve_id}`} className="hover:underline">
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
				</div>
			)}

			{/* Fixed Vulnerabilities */}
			{filteredFixedVulns.length > 0 && (
				<div className="card">
					<h2 className="text-xl font-bold text-green-900 mb-4">
						Fixed Vulnerabilities ({filteredFixedVulns.length})
					</h2>
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<SortableTableHeader
										label="CVE ID"
										sortKey="cve_id"
										currentSortKey={fixedVulnsSort.sortKey}
										currentSortDirection={fixedVulnsSort.sortDirection}
										onSort={fixedVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Package"
										sortKey="package_name"
										currentSortKey={fixedVulnsSort.sortKey}
										currentSortDirection={fixedVulnsSort.sortDirection}
										onSort={fixedVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Version"
										sortKey="package_version"
										currentSortKey={fixedVulnsSort.sortKey}
										currentSortDirection={fixedVulnsSort.sortDirection}
										onSort={fixedVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Type"
										sortKey="package_type"
										currentSortKey={fixedVulnsSort.sortKey}
										currentSortDirection={fixedVulnsSort.sortDirection}
										onSort={fixedVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Severity"
										sortKey="severity"
										currentSortKey={fixedVulnsSort.sortKey}
										currentSortDirection={fixedVulnsSort.sortDirection}
										onSort={fixedVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Status"
										sortKey="status"
										currentSortKey={fixedVulnsSort.sortKey}
										currentSortDirection={fixedVulnsSort.sortDirection}
										onSort={fixedVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="First Detected / Age"
										sortKey="first_detected_at"
										currentSortKey={fixedVulnsSort.sortKey}
										currentSortDirection={fixedVulnsSort.sortDirection}
										onSort={fixedVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="SLA Status"
										sortKey="sla_status"
										currentSortKey={fixedVulnsSort.sortKey}
										currentSortDirection={fixedVulnsSort.sortDirection}
										onSort={fixedVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Fix Version"
										sortKey="fix_version"
										currentSortKey={fixedVulnsSort.sortKey}
										currentSortDirection={fixedVulnsSort.sortDirection}
										onSort={fixedVulnsSort.handleSort}
									/>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{filteredFixedVulns.map((vuln) => {
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
												<Link to={`/vulnerabilities/${vuln.cve_id}`} className="hover:underline">
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
				</div>
			)}

			{/* Persistent Vulnerabilities */}
			{filteredPersistentVulns.length > 0 && (
				<div className="card">
					<h2 className="text-xl font-bold text-gray-900 mb-4">
						Persistent Vulnerabilities ({filteredPersistentVulns.length})
					</h2>
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<SortableTableHeader
										label="CVE ID"
										sortKey="cve_id"
										currentSortKey={persistentVulnsSort.sortKey}
										currentSortDirection={persistentVulnsSort.sortDirection}
										onSort={persistentVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Package"
										sortKey="package_name"
										currentSortKey={persistentVulnsSort.sortKey}
										currentSortDirection={persistentVulnsSort.sortDirection}
										onSort={persistentVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Version"
										sortKey="package_version"
										currentSortKey={persistentVulnsSort.sortKey}
										currentSortDirection={persistentVulnsSort.sortDirection}
										onSort={persistentVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Type"
										sortKey="package_type"
										currentSortKey={persistentVulnsSort.sortKey}
										currentSortDirection={persistentVulnsSort.sortDirection}
										onSort={persistentVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Severity"
										sortKey="severity"
										currentSortKey={persistentVulnsSort.sortKey}
										currentSortDirection={persistentVulnsSort.sortDirection}
										onSort={persistentVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Status"
										sortKey="status"
										currentSortKey={persistentVulnsSort.sortKey}
										currentSortDirection={persistentVulnsSort.sortDirection}
										onSort={persistentVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="First Detected / Age"
										sortKey="first_detected_at"
										currentSortKey={persistentVulnsSort.sortKey}
										currentSortDirection={persistentVulnsSort.sortDirection}
										onSort={persistentVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="SLA Status"
										sortKey="sla_status"
										currentSortKey={persistentVulnsSort.sortKey}
										currentSortDirection={persistentVulnsSort.sortDirection}
										onSort={persistentVulnsSort.handleSort}
									/>
									<SortableTableHeader
										label="Fix Version"
										sortKey="fix_version"
										currentSortKey={persistentVulnsSort.sortKey}
										currentSortDirection={persistentVulnsSort.sortDirection}
										onSort={persistentVulnsSort.handleSort}
									/>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{filteredPersistentVulns.map((vuln) => {
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
												<Link to={`/vulnerabilities/${vuln.cve_id}`} className="hover:underline">
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
				</div>
			)}

			{/* Vulnerability History Modal */}
			{historyVulnId && (
				<div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
					<div className="bg-white rounded-lg max-w-2xl w-full p-6 max-h-[80vh] overflow-y-auto">
						<div className="flex justify-between items-start mb-4">
							<h3 className="text-lg font-bold text-gray-900">Change History</h3>
							<button
								onClick={() => setHistoryVulnId(null)}
								className="text-gray-400 hover:text-gray-600"
							>
								✕
							</button>
						</div>
						<VulnerabilityHistory
							vulnerabilityId={historyVulnId}
							onClose={() => setHistoryVulnId(null)}
						/>
					</div>
				</div>
			)}
		</div>
	);
};
