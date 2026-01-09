import { FC, useEffect, useCallback, useState, useMemo } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useScan } from '../../hooks/useScans';
import { api } from '../../lib/api/client';
import { SeverityBadge } from '../ui/SeverityBadge';
import { StatusBadge } from '../ui/StatusBadge';
import { PackageCategoryBadge } from '../ui/PackageCategoryBadge';
import { VulnerabilityHistory } from '../ui/VulnerabilityHistory';
import { SortableTableHeader, useSortState } from '../ui/SortableTableHeader';
import { Pagination } from '../ui/Pagination';
import { formatDate, daysSince, calculateSLAStatus } from '../../lib/utils/formatters';
import type { Vulnerability } from '../../lib/api/types';

// Copyable text component with visual feedback
const CopyableText: FC<{ text: string; className?: string; mono?: boolean }> = ({ text, className = '', mono = false }) => {
	const [copied, setCopied] = useState(false);

	const handleCopy = async (e: React.MouseEvent) => {
		e.preventDefault();
		e.stopPropagation();

		try {
			await navigator.clipboard.writeText(text);
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		} catch (err) {
			console.error('Failed to copy:', err);
			// Fallback for older browsers or non-HTTPS contexts
			const textArea = document.createElement('textarea');
			textArea.value = text;
			textArea.style.position = 'fixed';
			textArea.style.left = '-999999px';
			document.body.appendChild(textArea);
			textArea.select();
			try {
				document.execCommand('copy');
				setCopied(true);
				setTimeout(() => setCopied(false), 2000);
			} catch (fallbackErr) {
				console.error('Fallback copy failed:', fallbackErr);
				alert('Failed to copy to clipboard');
			}
			document.body.removeChild(textArea);
		}
	};

	return (
		<button
			type="button"
			onClick={handleCopy}
			className={`group relative inline-flex items-center gap-2 cursor-pointer hover:bg-blue-50 active:bg-blue-100 rounded px-2 py-1 -mx-2 -my-1 transition-colors border-0 bg-transparent text-left ${className}`}
			title="Click to copy"
		>
			<span className={mono ? 'font-mono break-all' : ''}>{text}</span>
			{copied ? (
				<svg
					className="w-4 h-4 text-green-600 flex-shrink-0"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						strokeWidth={2}
						d="M5 13l4 4L19 7"
					/>
				</svg>
			) : (
				<svg
					className="w-4 h-4 text-gray-400 group-hover:text-blue-600 flex-shrink-0"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						strokeWidth={2}
						d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
					/>
				</svg>
			)}
			{copied && (
				<span className="absolute -top-8 left-1/2 transform -translate-x-1/2 bg-blue-600 text-white text-xs px-3 py-1.5 rounded shadow-lg whitespace-nowrap z-10">
					Copied!
				</span>
			)}
		</button>
	);
};

export const ScanDetails: FC = () => {
	const { id } = useParams<{ id: string }>();
	const navigate = useNavigate();
	const scanId = parseInt(id || '0', 10);
	const [currentPage, setCurrentPage] = useState(1);
	const itemsPerPage = 50;
	const [showUnfixable, setShowUnfixable] = useState(false);
	const [slaFilter, setSlaFilter] = useState<'all' | 'exceeded' | 'warning'>('all');
	const [historyVulnId, setHistoryVulnId] = useState<number | null>(null);
	const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
	const [bulkActionStatus, setBulkActionStatus] = useState<string>('');
	const [bulkActionNotes, setBulkActionNotes] = useState<string>('');
	const [showBulkActionModal, setShowBulkActionModal] = useState(false);
	const [bulkUpdating, setBulkUpdating] = useState(false);
	const { sortKey, sortDirection, handleSort } = useSortState('severity', 'desc');
	const { currentScan, loading, error, reload } = useScan(scanId, showUnfixable ? undefined : true);

	useEffect(() => {
		document.title = `Scan ${scanId} - Invulnerable`;
	}, [scanId]);

	// Reset to page 1 when filters change
	useEffect(() => {
		setCurrentPage(1);
	}, [showUnfixable, slaFilter]);

	const viewDiff = useCallback(() => {
		navigate(`/scans/${scanId}/diff`);
	}, [navigate, scanId]);

	const viewSBOM = useCallback(() => {
		navigate(`/scans/${scanId}/sbom`);
	}, [navigate, scanId]);

	const toggleSelection = useCallback((id: number) => {
		const newSelected = new Set(selectedIds);
		if (newSelected.has(id)) {
			newSelected.delete(id);
		} else {
			newSelected.add(id);
		}
		setSelectedIds(newSelected);
	}, [selectedIds]);

	const toggleSelectAll = (filteredVulns: Vulnerability[]) => {
		if (selectedIds.size === filteredVulns.length) {
			setSelectedIds(new Set());
		} else {
			setSelectedIds(new Set(filteredVulns.map(v => v.id)));
		}
	};

	const handleBulkUpdate = useCallback(async () => {
		if (selectedIds.size === 0) return;

		setBulkUpdating(true);
		try {
			await api.vulnerabilities.bulkUpdate(Array.from(selectedIds), {
				status: bulkActionStatus || undefined,
				notes: bulkActionNotes || undefined
			});
			reload();
			setSelectedIds(new Set());
			setShowBulkActionModal(false);
			setBulkActionStatus('');
			setBulkActionNotes('');
		} catch (e) {
			alert(e instanceof Error ? e.message : 'Failed to bulk update');
		} finally {
			setBulkUpdating(false);
		}
	}, [selectedIds, bulkActionStatus, bulkActionNotes, reload]);

	// Filter and sort vulnerabilities - must be called before conditional returns (Rules of Hooks)
	const filteredVulnerabilities = useMemo(() => {
		if (!currentScan) return [];

		const { scan, vulnerabilities } = currentScan;
		// Filter based on showUnfixable checkbox
		let filtered = showUnfixable
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
	}, [currentScan, showUnfixable, slaFilter, sortKey, sortDirection]);

	// Paginate the filtered vulnerabilities
	const paginatedVulnerabilities = useMemo(() => {
		const startIndex = (currentPage - 1) * itemsPerPage;
		const endIndex = startIndex + itemsPerPage;
		return filteredVulnerabilities.slice(startIndex, endIndex);
	}, [filteredVulnerabilities, currentPage, itemsPerPage]);

	const exportToCSV = useCallback(() => {
		if (!currentScan) return;

		const { scan, vulnerabilities } = currentScan;

		// Properly escape CSV values - handles quotes, newlines, and special characters
		const escapeCSV = (str: string) => {
			if (!str) return '';
			// Replace line breaks with spaces and escape quotes
			return str
				.replace(/\r?\n|\r/g, ' ') // Replace newlines with spaces
				.replace(/"/g, '""') // Escape quotes by doubling them
				.trim();
		};

		// Metadata rows at the top of the CSV
		const metadata = [
			['Scan Information'],
			['Scan ID', scan.id.toString()],
			['Image', scan.image_name],
			['Image Digest', scan.image_digest || 'N/A'],
			['Scan Date', formatDate(scan.scan_date)],
			['Syft Version', scan.syft_version || 'N/A'],
			['Grype Version', scan.grype_version || 'N/A'],
			['Total Vulnerabilities', vulnerabilities.length.toString()],
			['Filtered Vulnerabilities', filteredVulnerabilities.length.toString()],
			['SLA Limits', `Critical: ${scan.sla_critical}d, High: ${scan.sla_high}d, Medium: ${scan.sla_medium}d, Low: ${scan.sla_low}d`],
		];

		// CSV header
		const headers = [
			'CVE ID',
			'Severity',
			'Package Name',
			'Installed Version',
			'Fixed Version',
			'Status',
			'First Detected',
			'Days Since Detection',
			'SLA Days',
			'Days Remaining',
			'SLA Status',
			'Notes',
			'Description',
			'URL'
		];

		// Number of columns in the data table
		const numColumns = headers.length;

		// Pad metadata rows to match the number of columns in the data table
		const paddedMetadata = metadata.map(row => {
			const padded = [...row];
			while (padded.length < numColumns) {
				padded.push('');
			}
			return padded;
		});

		// Convert vulnerabilities to CSV rows
		const rows = filteredVulnerabilities.map(vuln => {
			const daysSinceDetection = daysSince(vuln.first_detected_at);

			// Get SLA limit based on severity
			let slaDays: number;
			switch (vuln.severity) {
				case 'Critical':
					slaDays = scan.sla_critical;
					break;
				case 'High':
					slaDays = scan.sla_high;
					break;
				case 'Medium':
					slaDays = scan.sla_medium;
					break;
				case 'Low':
					slaDays = scan.sla_low;
					break;
				default:
					slaDays = 0;
			}

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

			return [
				vuln.cve_id,
				vuln.severity,
				vuln.package_name,
				vuln.package_version,
				vuln.fix_version || 'N/A',
				vuln.status,
				formatDate(vuln.first_detected_at),
				daysSinceDetection.toString(),
				slaDays.toString(),
				slaStatus.daysRemaining.toString(),
				slaStatus.status,
				vuln.notes || '',
				vuln.description || '',
				vuln.url || ''
			];
		});

		// Combine metadata, headers, and rows with proper CSV escaping
		const csvContent = [
			// Add padded metadata rows (escape content, then wrap in quotes)
			...paddedMetadata
				.filter(row => row.length > 0)
				.map(row => row.map(cell => `"${escapeCSV(cell)}"`).join(',')),
			// Empty line separator (also pad to match column count)
			Array(numColumns).fill('""').join(','),
			// Add vulnerability data table (escape headers and data)
			headers.map(h => `"${escapeCSV(h)}"`).join(','),
			...rows.map(row => row.map(cell => `"${escapeCSV(cell)}"`).join(','))
		].join('\n');

		// Generate safe filename from image name
		// Parse image name to extract name and tag (format: registry/repo:tag or repo:tag)
		const imageParts = scan.image_name.split(':');
		const tag = imageParts.length > 1 ? imageParts[imageParts.length - 1] : 'latest';
		const imageNamePart = imageParts.slice(0, -1).join(':') || imageParts[0];

		// Extract just the repo name (remove registry if present)
		const repoName = imageNamePart.split('/').pop() || imageNamePart;

		// Sanitize for filesystem safety (remove/replace invalid characters)
		// Invalid characters: < > : " / \ | ? * and control characters
		const sanitize = (str: string) => str
			.replace(/[<>:"/\\|?*\x00-\x1f]/g, '_') // Replace invalid chars with underscore
			.replace(/\s+/g, '_') // Replace whitespace with underscore
			.replace(/_{2,}/g, '_') // Replace multiple underscores with single
			.replace(/^_+|_+$/g, ''); // Trim underscores from start/end

		const safeRepoName = sanitize(repoName);
		const safeTag = sanitize(tag);
		const dateStr = new Date().toISOString().split('T')[0];

		const filename = `${safeRepoName}_${safeTag}-${dateStr}.csv`;

		// Create blob and download
		const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.setAttribute('href', url);
		link.setAttribute('download', filename);
		link.style.visibility = 'hidden';
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
	}, [currentScan, filteredVulnerabilities, scanId]);

	const { scan, vulnerabilities } = currentScan || { scan: null, vulnerabilities: [] };

	return (
		<div className="space-y-6">
			<div className="flex justify-between items-center">
				<h1 className="text-3xl font-bold text-gray-900">Scan #{scan?.id || scanId}</h1>
				<div className="space-x-2">
					<button
						onClick={exportToCSV}
						className="btn btn-secondary"
						disabled={!scan}
						title="Export filtered vulnerabilities to CSV"
					>
						<svg className="w-4 h-4 mr-1 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
						</svg>
						Export CSV
					</button>
					<button onClick={viewSBOM} className="btn btn-secondary" disabled={!scan}>
						View SBOM
					</button>
					<button onClick={viewDiff} className="btn btn-primary" disabled={!scan}>
						View Diff
					</button>
				</div>
			</div>

			{/* Scan Details */}
			{scan && (
				<>
					<div className="card">
						<h2 className="text-xl font-bold text-gray-900 mb-4">Scan Details</h2>
						<dl className="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<dt className="text-sm font-medium text-gray-500">Image</dt>
								<dd className="mt-1 text-sm text-gray-900">
									<CopyableText text={scan.image_name} />
								</dd>
							</div>
							<div>
								<dt className="text-sm font-medium text-gray-500">Image Digest</dt>
								<dd className="mt-1 text-sm text-gray-900">
									{scan.image_digest ? (
										<CopyableText text={scan.image_digest} mono={true} />
									) : (
										<span className="text-gray-500">N/A</span>
									)}
								</dd>
							</div>
							<div>
								<dt className="text-sm font-medium text-gray-500">Scan Date</dt>
								<dd className="mt-1 text-sm text-gray-900">
									<CopyableText text={formatDate(scan.scan_date)} />
								</dd>
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
				</>
			)}

			{/* Bulk action bar */}
			{selectedIds.size > 0 && (
				<div className="bg-blue-50 border border-blue-200 rounded-lg p-4 flex justify-between items-center">
					<span className="text-sm font-medium text-blue-900">
						{selectedIds.size} vulnerabilit{selectedIds.size === 1 ? 'y' : 'ies'} selected
					</span>
					<div className="space-x-2">
						<button
							onClick={() => setShowBulkActionModal(true)}
							className="btn btn-primary"
						>
							Update Selected
						</button>
						<button
							onClick={() => setSelectedIds(new Set())}
							className="btn btn-secondary"
						>
							Clear Selection
						</button>
					</div>
				</div>
			)}

			{/* Vulnerabilities List */}
			<div className="card">
				<div className="flex justify-between items-center mb-4">
					<h2 className="text-xl font-bold text-gray-900">Vulnerabilities</h2>
					<div className="flex items-center space-x-4">
						<label className="flex items-center space-x-2 text-sm">
							<input
								type="checkbox"
								checked={showUnfixable}
								onChange={(e) => setShowUnfixable(e.target.checked)}
								className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
							/>
							<span className="text-gray-700">Show unfixable CVEs</span>
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

				{error && (
					<div className="bg-red-50 p-4 rounded mb-4">
						<p className="text-red-600">{error}</p>
					</div>
				)}

				{!loading && (
					<p className="text-sm text-gray-600 mb-4">
						Showing {filteredVulnerabilities.length} of {vulnerabilities.length} vulnerabilities
					</p>
				)}

				{loading ? (
					<div className="flex items-center justify-center py-32">
						<div className="flex flex-col items-center gap-3">
							<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
							<p className="text-gray-500 text-sm">Loading vulnerabilities...</p>
						</div>
					</div>
				) : filteredVulnerabilities.length === 0 ? (
					<p className="text-gray-500">No vulnerabilities found</p>
				) : scan ? (
					<div className="overflow-x-auto">
						<table className="min-w-full divide-y divide-gray-200">
							<thead className="bg-gray-50">
								<tr>
									<th className="px-6 py-3 text-left">
										<input
											type="checkbox"
											checked={selectedIds.size === filteredVulnerabilities.length && filteredVulnerabilities.length > 0}
											onChange={() => toggleSelectAll(filteredVulnerabilities)}
											className="rounded border-gray-300"
										/>
									</th>
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
								{paginatedVulnerabilities.map((vuln) => {
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
											<td className="px-6 py-4">
												<input
													type="checkbox"
													checked={selectedIds.has(vuln.id)}
													onChange={() => toggleSelection(vuln.id)}
													className="rounded border-gray-300"
												/>
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
				) : null}

				{/* Pagination */}
				{filteredVulnerabilities.length > 0 && (
					<Pagination
						currentPage={currentPage}
						totalItems={filteredVulnerabilities.length}
						itemsPerPage={itemsPerPage}
						onPageChange={setCurrentPage}
					/>
				)}
			</div>

			{/* Bulk action modal */}
			{showBulkActionModal && (
				<div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
					<div className="bg-white rounded-lg max-w-md w-full p-6">
						<h3 className="text-lg font-bold mb-4">
							Update {selectedIds.size} Vulnerabilit{selectedIds.size === 1 ? 'y' : 'ies'}
						</h3>
						<div className="space-y-4">
							<div>
								<label className="block text-sm font-medium text-gray-700 mb-1">
									Status
								</label>
								<select
									value={bulkActionStatus}
									onChange={(e) => setBulkActionStatus(e.target.value)}
									className="w-full rounded-md border-gray-300"
								>
									<option value="">-- No change --</option>
									<option value="active">Active</option>
									<option value="in_progress">In Progress</option>
									<option value="fixed">Fixed</option>
									<option value="ignored">Ignored</option>
									<option value="accepted">Accepted</option>
								</select>
							</div>
							<div>
								<label className="block text-sm font-medium text-gray-700 mb-1">
									{bulkActionStatus === 'accepted' ? 'Acceptance Justification' : 'Notes'}
								</label>
								<textarea
									value={bulkActionNotes}
									onChange={(e) => setBulkActionNotes(e.target.value)}
									rows={3}
									placeholder={bulkActionStatus === 'accepted'
										? 'Explain why this risk is being accepted...'
										: 'Optional notes about this update...'}
									className="w-full rounded-md border-gray-300"
								/>
							</div>
						</div>
						<div className="mt-6 flex justify-end space-x-2">
							<button
								onClick={() => setShowBulkActionModal(false)}
								className="btn btn-secondary"
								disabled={bulkUpdating}
							>
								Cancel
							</button>
							<button
								onClick={handleBulkUpdate}
								className="btn btn-primary"
								disabled={bulkUpdating || (!bulkActionStatus && !bulkActionNotes)}
							>
								{bulkUpdating ? 'Updating...' : 'Update'}
							</button>
						</div>
					</div>
				</div>
			)}

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
