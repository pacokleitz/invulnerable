import { FC, useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../../lib/api/client';
import type { ScanDiff as ScanDiffType, Scan } from '../../lib/api/types';
import { SeverityBadge } from '../ui/SeverityBadge';
import { StatusBadge } from '../ui/StatusBadge';
import { PackageCategoryBadge } from '../ui/PackageCategoryBadge';
import { VulnerabilityHistory } from '../ui/VulnerabilityHistory';
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

	// Filter vulnerabilities based on showUnfixed
	const filteredNewVulns = showUnfixed
		? diff?.new_vulnerabilities || []
		: (diff?.new_vulnerabilities || []).filter(v => v.fix_version !== null && v.fix_version !== undefined);

	const filteredFixedVulns = showUnfixed
		? diff?.fixed_vulnerabilities || []
		: (diff?.fixed_vulnerabilities || []).filter(v => v.fix_version !== null && v.fix_version !== undefined);

	const filteredPersistentVulns = showUnfixed
		? diff?.persistent_vulnerabilities || []
		: (diff?.persistent_vulnerabilities || []).filter(v => v.fix_version !== null && v.fix_version !== undefined);

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
										Type
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
										vuln.remediation_date
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
										Type
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
										vuln.remediation_date
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
										Type
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
										vuln.remediation_date
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
