import { FC, useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../../lib/api/client';
import type { ScanDiff as ScanDiffType } from '../../lib/api/types';
import { SeverityBadge } from '../ui/SeverityBadge';
import { StatusBadge } from '../ui/StatusBadge';
import { formatDate } from '../../lib/utils/formatters';

export const ScanDiff: FC = () => {
	const { id } = useParams<{ id: string }>();
	const scanId = parseInt(id || '0', 10);
	const [diff, setDiff] = useState<ScanDiffType | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		document.title = `Scan Diff - Scan ${scanId} - Invulnerable`;

		const fetchDiff = async () => {
			try {
				const data = await api.scans.getDiff(scanId);
				setDiff(data);
			} catch (e) {
				setError(e instanceof Error ? e.message : 'Failed to load scan diff');
			} finally {
				setLoading(false);
			}
		};

		fetchDiff();
	}, [scanId]);

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

	if (!diff) {
		return null;
	}

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<h1 className="text-3xl font-bold text-gray-900">Scan Comparison</h1>
				<Link to={`/scans/${scanId}`} className="text-blue-600 hover:text-blue-800">
					← Back to Scan
				</Link>
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
			{diff.new_vulnerabilities.length > 0 && (
				<div className="card">
					<h2 className="text-xl font-bold text-red-900 mb-4">
						New Vulnerabilities ({diff.new_vulnerabilities.length})
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
										Severity
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Status
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										First Detected
									</th>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{diff.new_vulnerabilities.map((vuln) => (
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
											<SeverityBadge severity={vuln.severity} />
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm">
											<StatusBadge status={vuln.status} />
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{formatDate(vuln.first_detected_at)}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				</div>
			)}

			{/* Fixed Vulnerabilities */}
			{diff.fixed_vulnerabilities.length > 0 && (
				<div className="card">
					<h2 className="text-xl font-bold text-green-900 mb-4">
						Fixed Vulnerabilities ({diff.fixed_vulnerabilities.length})
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
										Severity
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										First Detected
									</th>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{diff.fixed_vulnerabilities.map((vuln) => (
									<tr key={vuln.id} className="hover:bg-gray-50">
										<td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-blue-600">
											{vuln.cve_id}
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
										<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{formatDate(vuln.first_detected_at)}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				</div>
			)}

			{/* Persistent Vulnerabilities */}
			{diff.persistent_vulnerabilities.length > 0 && (
				<div className="card">
					<h2 className="text-xl font-bold text-gray-900 mb-4">
						Persistent Vulnerabilities ({diff.persistent_vulnerabilities.length})
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
										Severity
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										Status
									</th>
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										First Detected
									</th>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{diff.persistent_vulnerabilities.map((vuln) => (
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
											<SeverityBadge severity={vuln.severity} />
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm">
											<StatusBadge status={vuln.status} />
										</td>
										<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{formatDate(vuln.first_detected_at)}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				</div>
			)}
		</div>
	);
};
