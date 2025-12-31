import { FC, useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../../lib/api/client';
import type { Vulnerability } from '../../lib/api/types';
import { SeverityBadge } from '../ui/SeverityBadge';
import { StatusBadge } from '../ui/StatusBadge';
import { VulnerabilityHistory } from '../ui/VulnerabilityHistory';
import { formatDate, daysSince, calculateSLAStatus } from '../../lib/utils/formatters';

export const CVEDetails: FC = () => {
	const { cve } = useParams<{ cve: string }>();
	const [vulnerabilities, setVulnerabilities] = useState<Vulnerability[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [historyVulnId, setHistoryVulnId] = useState<number | null>(null);

	const loadVulnerabilities = useCallback(async () => {
		if (!cve) return;
		setLoading(true);
		setError(null);
		try {
			// Use the list endpoint with cve_id filter to get image context
			const data = await api.vulnerabilities.list({
				cve_id: cve,
				limit: 200
			});
			setVulnerabilities(data);
		} catch (e) {
			setError(e instanceof Error ? e.message : 'Failed to load vulnerability');
		} finally {
			setLoading(false);
		}
	}, [cve]);

	useEffect(() => {
		document.title = `${cve} - Invulnerable`;
		loadVulnerabilities();
	}, [cve, loadVulnerabilities]);

	if (loading) {
		return (
			<div className="text-center py-12">
				<p className="text-gray-500">Loading vulnerability details...</p>
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

	if (vulnerabilities.length === 0) {
		return (
			<div className="card text-center py-12">
				<p className="text-gray-500">Vulnerability not found</p>
			</div>
		);
	}

	const cveInfo = vulnerabilities[0];

	return (
		<>
			<div className="space-y-6">
				<div className="flex justify-between items-center">
					<h1 className="text-3xl font-bold text-gray-900">{cve}</h1>
					<Link to="/vulnerabilities" className="text-blue-600 hover:text-blue-800">
						← Back to Vulnerabilities
					</Link>
				</div>

				{/* CVE Summary */}
				<div className="card">
					<h2 className="text-xl font-bold text-gray-900 mb-4">Vulnerability Details</h2>
					<dl className="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<dt className="text-sm font-medium text-gray-500">CVE ID</dt>
							<dd className="mt-1 text-sm text-gray-900">{cveInfo.cve_id}</dd>
						</div>
						<div>
							<dt className="text-sm font-medium text-gray-500">Severity</dt>
							<dd className="mt-1">
								<SeverityBadge severity={cveInfo.severity} />
							</dd>
						</div>
						<div>
							<dt className="text-sm font-medium text-gray-500">Description</dt>
							<dd className="mt-1 text-sm text-gray-900">{cveInfo.description || 'N/A'}</dd>
						</div>
						<div>
							<dt className="text-sm font-medium text-gray-500">Fix Available</dt>
							<dd className="mt-1 text-sm text-gray-900">{cveInfo.fix_version || 'No fix available'}</dd>
						</div>
						<div className="md:col-span-2">
							<dt className="text-sm font-medium text-gray-500">Affected Images</dt>
							<dd className="mt-1 text-sm text-gray-900">{vulnerabilities.length}</dd>
						</div>
					</dl>
				</div>

				{/* Affected Images */}
				<div className="card">
					<h2 className="text-xl font-bold text-gray-900 mb-4">Affected Images</h2>
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Image
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Package
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Version
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
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{vulnerabilities.map((vuln) => {
									const slaStatus = calculateSLAStatus(
										vuln.first_detected_at_for_image || vuln.first_detected_at,
										vuln.severity,
										{
											critical: vuln.sla_critical || 7,
											high: vuln.sla_high || 30,
											medium: vuln.sla_medium || 90,
											low: vuln.sla_low || 180,
										},
										vuln.status,
										vuln.remediation_date,
										vuln.updated_at
									);

									return (
										<tr key={`${vuln.id}-${vuln.image_id}`} className="hover:bg-gray-50">
											<td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-blue-600">
												<Link
													to={`/scans/${vuln.latest_scan_id}`}
													className="hover:underline"
												>
													{vuln.image_name}
												</Link>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
												{vuln.package_name}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
												{vuln.package_version}
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm">
												<StatusBadge
													status={vuln.status}
													onClick={() => setHistoryVulnId(vuln.id)}
												/>
											</td>
											<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
												<div>
													{formatDate(vuln.first_detected_at_for_image || vuln.first_detected_at)}
												</div>
												<div className="text-xs text-gray-400">
													({daysSince(vuln.first_detected_at_for_image || vuln.first_detected_at)} days ago)
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
										</tr>
									);
								})}
							</tbody>
						</table>
					</div>
				</div>
			</div>

			{/* Vulnerability History Modal */}
			{historyVulnId && (
				<VulnerabilityHistory
					vulnerabilityId={historyVulnId}
					onClose={() => setHistoryVulnId(null)}
				/>
			)}
		</>
	);
};
