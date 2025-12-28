import { FC, useEffect, useCallback, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useScan } from '../../hooks/useScans';
import { SeverityBadge } from '../ui/SeverityBadge';
import { StatusBadge } from '../ui/StatusBadge';
import { formatDate, daysSince } from '../../lib/utils/formatters';

export const ScanDetails: FC = () => {
	const { id } = useParams<{ id: string }>();
	const navigate = useNavigate();
	const scanId = parseInt(id || '0', 10);
	const [showUnfixed, setShowUnfixed] = useState(false);
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

	// Filter vulnerabilities based on showUnfixed checkbox
	const filteredVulnerabilities = showUnfixed
		? vulnerabilities
		: vulnerabilities.filter(vuln => vuln.fix_version !== null && vuln.fix_version !== undefined);

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
							<p className="text-sm font-medium text-red-800">Critical</p>
							<p className="text-2xl font-bold text-red-900">{scan.critical_count}</p>
						</div>
					</div>
					<div className="flex items-center justify-between p-4 bg-orange-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-orange-800">High</p>
							<p className="text-2xl font-bold text-orange-900">{scan.high_count}</p>
						</div>
					</div>
					<div className="flex items-center justify-between p-4 bg-yellow-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-yellow-800">Medium</p>
							<p className="text-2xl font-bold text-yellow-900">{scan.medium_count}</p>
						</div>
					</div>
					<div className="flex items-center justify-between p-4 bg-blue-50 rounded-lg">
						<div>
							<p className="text-sm font-medium text-blue-800">Low</p>
							<p className="text-2xl font-bold text-blue-900">{scan.low_count}</p>
						</div>
					</div>
				</div>
			</div>

			{/* Vulnerabilities List */}
			<div className="card">
				<div className="flex justify-between items-center mb-4">
					<h2 className="text-xl font-bold text-gray-900">Vulnerabilities</h2>
					<label className="flex items-center space-x-2 text-sm">
						<input
							type="checkbox"
							checked={showUnfixed}
							onChange={(e) => setShowUnfixed(e.target.checked)}
							className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
						/>
						<span className="text-gray-700">Show unfixed CVEs</span>
					</label>
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
										Fix Version
									</th>
								</tr>
							</thead>
							<tbody className="bg-white divide-y divide-gray-200">
								{filteredVulnerabilities.map((vuln) => (
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
										<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{vuln.fix_version || 'N/A'}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				)}
			</div>
		</div>
	);
};
