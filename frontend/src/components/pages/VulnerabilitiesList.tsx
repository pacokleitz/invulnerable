import { FC, useState, useEffect, useCallback } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { api } from '../../lib/api/client';
import type { Vulnerability } from '../../lib/api/types';
import { SeverityBadge } from '../ui/SeverityBadge';
import { StatusBadge } from '../ui/StatusBadge';
import { PackageCategoryBadge } from '../ui/PackageCategoryBadge';
import { VulnerabilityHistory } from '../ui/VulnerabilityHistory';
import { SortableTableHeader, useSortState } from '../ui/SortableTableHeader';
import { formatDate, daysSince, calculateSLAStatus } from '../../lib/utils/formatters';
import { categorizePackageType } from '../../lib/utils/packageTypes';

export const VulnerabilitiesList: FC = () => {
	const [vulnerabilities, setVulnerabilities] = useState<Vulnerability[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [searchParams, setSearchParams] = useSearchParams();
	const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
	const [bulkActionStatus, setBulkActionStatus] = useState<string>('');
	const [bulkActionNotes, setBulkActionNotes] = useState<string>('');
	const [showBulkActionModal, setShowBulkActionModal] = useState(false);
	const [bulkUpdating, setBulkUpdating] = useState(false);
	const [historyVulnId, setHistoryVulnId] = useState<number | null>(null);

	// Sorting state
	const { sortKey, sortDirection, handleSort } = useSortState('first_detected_at', 'desc');

	// Filters from URL
	const severityFilter = searchParams.get('severity') || '';
	const statusFilter = searchParams.get('status') || '';
	const imageFilter = searchParams.get('image') || '';
	const cveFilter = searchParams.get('cve') || '';
	const packageCategoryFilter = searchParams.get('package_category') || '';
	const showUnfixed = searchParams.get('show_unfixed') === 'true'; // Default to false

	const loadVulnerabilities = useCallback(async () => {
		setLoading(true);
		setError(null);

		try {
			const params: { limit: number; severity?: string; status?: string; image_name?: string; cve_id?: string; has_fix?: boolean } = { limit: 200 };
			if (severityFilter) params.severity = severityFilter;
			if (statusFilter) params.status = statusFilter;
			if (imageFilter) params.image_name = imageFilter;
			if (cveFilter) params.cve_id = cveFilter;
			// When showUnfixed is false, only show CVEs with fixes (has_fix = true)
			if (!showUnfixed) params.has_fix = true;

			let data = await api.vulnerabilities.list(params);

			// Apply client-side package category filter
			if (packageCategoryFilter) {
				data = data.filter(vuln => {
					const category = categorizePackageType(vuln.package_type).category;
					return category === packageCategoryFilter;
				});
			}

			// Apply client-side sorting
			if (sortKey && sortDirection) {
				data.sort((a, b) => {
					let aVal: any = a[sortKey as keyof Vulnerability];
					let bVal: any = b[sortKey as keyof Vulnerability];

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

					// Compare
					if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
					if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
					return 0;
				});
			}

			setVulnerabilities(data);
		} catch (e) {
			setError(e instanceof Error ? e.message : 'Failed to load vulnerabilities');
		} finally {
			setLoading(false);
		}
	}, [severityFilter, statusFilter, imageFilter, cveFilter, packageCategoryFilter, showUnfixed, sortKey, sortDirection]);

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

	const toggleSelection = (id: number) => {
		const newSelected = new Set(selectedIds);
		if (newSelected.has(id)) {
			newSelected.delete(id);
		} else {
			newSelected.add(id);
		}
		setSelectedIds(newSelected);
	};

	const toggleSelectAll = () => {
		if (selectedIds.size === vulnerabilities.length) {
			setSelectedIds(new Set());
		} else {
			setSelectedIds(new Set(vulnerabilities.map(v => v.id)));
		}
	};

	const handleBulkUpdate = async () => {
		if (selectedIds.size === 0) return;

		setBulkUpdating(true);
		try {
			await api.vulnerabilities.bulkUpdate(Array.from(selectedIds), {
				status: bulkActionStatus || undefined,
				notes: bulkActionNotes || undefined
			});
			await loadVulnerabilities();
			setSelectedIds(new Set());
			setShowBulkActionModal(false);
			setBulkActionStatus('');
			setBulkActionNotes('');
		} catch (e) {
			alert(e instanceof Error ? e.message : 'Failed to bulk update');
		} finally {
			setBulkUpdating(false);
		}
	};

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
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
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
							<option value="in_progress">In Progress</option>
							<option value="fixed">Fixed</option>
							<option value="ignored">Ignored</option>
							<option value="accepted">Accepted</option>
						</select>
					</div>

					<div>
						<label htmlFor="packageCategory" className="block text-sm font-medium text-gray-700">
							Package Type
						</label>
						<select
							id="packageCategory"
							value={packageCategoryFilter}
							onChange={(e) => updateFilter('package_category', e.target.value)}
							className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						>
							<option value="">All</option>
							<option value="os">OS/System</option>
							<option value="application">App/Library</option>
							<option value="unknown">Unknown</option>
						</select>
					</div>

					<div>
						<label htmlFor="imageFilter" className="block text-sm font-medium text-gray-700">
							Image Name
						</label>
						<input
							type="text"
							id="imageFilter"
							value={imageFilter}
							onChange={(e) => updateFilter('image', e.target.value)}
							placeholder="e.g., nginx:latest"
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
									<th className="px-6 py-3 text-left">
										<input
											type="checkbox"
											checked={selectedIds.size === vulnerabilities.length && vulnerabilities.length > 0}
											onChange={toggleSelectAll}
											className="rounded border-gray-300"
										/>
									</th>
									<SortableTableHeader
										label="Image"
										sortKey="image_name"
										currentSortKey={sortKey}
										currentSortDirection={sortDirection}
										onSort={handleSort}
									/>
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
									<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
										SLA Status
									</th>
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
												},
												vuln.status,
												vuln.remediation_date
										  )
										: null;

									return (
										<tr key={`${vuln.id}-${vuln.image_id || 'unknown'}`} className="hover:bg-gray-50">
											<td className="px-6 py-4">
												<input
													type="checkbox"
													checked={selectedIds.has(vuln.id)}
													onChange={() => toggleSelection(vuln.id)}
													className="rounded border-gray-300"
												/>
											</td>
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
												{slaStatus ? (
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
